package store

import (
	"context"
	"fmt"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

func playSave() error {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"localhost:2379"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return fmt.Errorf("error creating client: %w", err)
	}

	defer cli.Close()

	ctx := context.TODO()

	const fakeKey = "calculations/FibonacciOf/28h48h4994h"

	// Get the key.
	getResp, err := cli.Get(ctx, "calculations/FibonacciOf/28h48h4994h")
	if err != nil {
		return fmt.Errorf("could not get key: %w", err)
	}

	if getResp.Count != 1 {
		return fmt.Errorf("expected 1 key got got (%d)", getResp.Count)
	}
	kv := getResp.Kvs[0]

	// Now pretend we update it...
	kv.Value = []byte("serialization of done with result probably")

	txResp, err := cli.Txn(context.TODO()).If(
		clientv3.Compare(clientv3.Version(fakeKey), "=", kv.Version),
		// OR it doesn't exist??
	).Then(
		clientv3.OpPut(fakeKey, "the serialized updated model"),
	).Commit()
	if err != nil {
		return fmt.Errorf("error during transaction: %w", err)
	} else if !txResp.Succeeded {
		return fmt.Errorf("transaction did not succeed")
	}

	return nil
}

type etcdClient interface {
	Get(context.Context, string, ...clientv3.OpOption) (*clientv3.GetResponse, error)
	Txn(context.Context) clientv3.Txn
	Then(...clientv3.Op) clientv3.Txn
	Commit() (*clientv3.TxnResponse, error)
}

func (c Calculation) Save(ctx context.Context, store etcdClient) error {
	return nil
}
