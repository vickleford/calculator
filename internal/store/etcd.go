package store

import (
	"context"
	"encoding/json"
	"fmt"

	clientv3 "go.etcd.io/etcd/client/v3"
)

type etcdClient interface {
	Get(context.Context, string, ...clientv3.OpOption) (*clientv3.GetResponse, error)
	Put(context.Context, string, string, ...clientv3.OpOption) (*clientv3.PutResponse, error)
	Txn(context.Context) clientv3.Txn
}

// Save creates or updates a Calculation in etcd.
func (c Calculation) Save(ctx context.Context, store etcdClient) error {
	value, err := json.Marshal(c)
	if err != nil {
		return fmt.Errorf("unable to marshal calculation %q to JSON: %w", c.Name, err)
	}

	getResp, err := store.Get(ctx, c.Key())
	if err != nil {
		return fmt.Errorf("unable to get key: %w", err)
	}

	if getResp.Count == 0 {
		_, err := store.Put(ctx, c.Key(), string(value))
		if err != nil {
			return fmt.Errorf("error writing key %q: %w", c.Key(), err)
		}

		return nil
	}

	if getResp.Count > 1 {
		return fmt.Errorf("key %q is a prefix", c.Key())
	}

	kv := getResp.Kvs[0]

	txResp, err := store.Txn(ctx).If(
		clientv3.Compare(clientv3.Version(c.Key()), "=", kv.Version),
	).Then(
		clientv3.OpPut(c.Key(), string(value)),
	).Commit()
	if err != nil {
		return fmt.Errorf("transaction error: %w", err)
	} else if !txResp.Succeeded {
		return fmt.Errorf("transaction did not succeed")
	}

	return nil
}

func (c Calculation) Key() string {
	return fmt.Sprintf("calculations/%s", c.Name)
}
