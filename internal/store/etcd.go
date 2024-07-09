package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	clientv3 "go.etcd.io/etcd/client/v3"
)

var ErrKeyAlreadyExists = errors.New("key already exists")

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

func CalculationKey(calculation Calculation) string {
	return fmt.Sprintf("calculations/%s", calculation.Name)
}
