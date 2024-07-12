package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

var (
	ErrKeyAlreadyExists   = errors.New("key already exists")
	ErrKeyNotFound        = errors.New("key not found")
	ErrUpdateUnsuccessful = errors.New("update was not successful")
)

type etcdClient interface {
	Get(context.Context, string, ...clientv3.OpOption) (*clientv3.GetResponse, error)
	Put(context.Context, string, string, ...clientv3.OpOption) (*clientv3.PutResponse, error)
	Txn(context.Context) clientv3.Txn
}

type CalculationStore struct {
	cli etcdClient
}

func NewCalculationStore(cli etcdClient) *CalculationStore {
	return &CalculationStore{cli: cli}
}

// Create creates a Calculation for the first time or errors. If the key already
// exists, it returns ErrKeyAlreadyExist.
func (c *CalculationStore) Create(ctx context.Context, calculation Calculation) error {
	key := CalculationKey(calculation)

	value, err := json.Marshal(calculation)
	if err != nil {
		return fmt.Errorf("unable to marshal calculation %q to JSON: %w", calculation.Name, err)
	}

	resp, err := c.cli.Txn(ctx).If(
		clientv3.Compare(clientv3.CreateRevision(key), "=", 0),
	).Then(
		clientv3.OpPut(key, string(value)),
	).Commit()
	if err != nil {
		return fmt.Errorf("error writing key: %q: %w", key, err)
	}

	if !resp.Succeeded {
		return ErrKeyAlreadyExists
	}

	return nil
}

// SetStartedAt records the Started time for the Calculation with the given name.
func (c *CalculationStore) SetStartedTime(ctx context.Context, name string, t time.Time) error {
	calculation := Calculation{Name: name}
	key := CalculationKey(calculation)

	getResp, err := c.cli.Get(ctx, key)
	if err != nil {
		return fmt.Errorf("unable to get calculation: %w", err)
	}
	if getResp.Count == 0 {
		return fmt.Errorf("calculation %q does not exist", name)
	}
	if getResp.Count > 1 {
		return fmt.Errorf("key %q is a prefix", key)
	}

	if err := json.Unmarshal(getResp.Kvs[0].Value, &calculation); err != nil {
		return fmt.Errorf("error unmarshaling calculation %q: %s", name, err)
	}
	calculation.Metadata.Started = &t

	update, err := json.Marshal(calculation)
	if err != nil {
		return fmt.Errorf("error marshaling updated calculation %q: %s", name, err)
	}

	resp, err := c.cli.Txn(ctx).If(
		clientv3.Compare(clientv3.Version(key), "=", getResp.Kvs[0].Version),
	).Then(
		clientv3.OpPut(key, string(update)),
	).Commit()
	if err != nil {
		return fmt.Errorf("error writing key %q: %w", key, err)
	}

	if !resp.Succeeded {
		return ErrUpdateUnsuccessful
	}

	return nil
}

// Save saves or updates a Calculation in etcd.
func (c *CalculationStore) Save(ctx context.Context, calculation Calculation) error {
	key := CalculationKey(calculation)

	value, err := json.Marshal(calculation)
	if err != nil {
		return fmt.Errorf("unable to marshal calculation %q to JSON: %w", calculation.Name, err)
	}

	getResp, err := c.cli.Get(ctx, key)
	if err != nil {
		return fmt.Errorf("unable to get key: %w", err)
	}

	if getResp.Count == 0 {
		_, err := c.cli.Put(ctx, key, string(value))
		if err != nil {
			return fmt.Errorf("error writing key %q: %w", key, err)
		}

		return nil
	}

	if getResp.Count > 1 {
		return fmt.Errorf("key %q is a prefix", key)
	}

	kv := getResp.Kvs[0]

	txResp, err := c.cli.Txn(ctx).If(
		clientv3.Compare(clientv3.Version(key), "=", kv.Version),
	).Then(
		clientv3.OpPut(key, string(value)),
	).Commit()
	if err != nil {
		return fmt.Errorf("transaction error: %w", err)
	} else if !txResp.Succeeded {
		return fmt.Errorf("transaction did not succeed")
	}

	return nil
}

// Get returns the calculation for the given name.
func (c *CalculationStore) Get(ctx context.Context, name string) (Calculation, error) {
	calc := Calculation{Name: name}

	key := CalculationKey(calc)

	getResp, err := c.cli.Get(ctx, key)
	if err != nil {
		return calc, fmt.Errorf("error getting key %q: %w", key, err)
	}

	if getResp.Count == 0 {
		return calc, ErrKeyNotFound
	}

	if getResp.Count != 1 {
		return calc, fmt.Errorf("ambiguous results: %d keys found", getResp.Count)
	}

	if err := json.Unmarshal(getResp.Kvs[0].Value, &calc); err != nil {
		return calc, fmt.Errorf("error unmarshaling calculation: %w", err)
	}

	return calc, err
}

func CalculationKey(calculation Calculation) string {
	return fmt.Sprintf("calculations/%s", calculation.Name)
}
