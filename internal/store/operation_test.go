package store_test

import (
	"context"
	"testing"
	"time"

	"github.com/vickleford/calculator/internal/store"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type etcdClientSpy struct {
	ComparisonsSeenByIf  []clientv3.Cmp
	OperationsSeenByThen []clientv3.Op
	ReturnTxnResponse    *clientv3.TxnResponse
	CommitError          error

	ShouldTxnIfSucceed bool
}

func NewETCDClientSpy() *etcdClientSpy {
	return &etcdClientSpy{
		ComparisonsSeenByIf:  make([]clientv3.Cmp, 0),
		OperationsSeenByThen: make([]clientv3.Op, 0),
	}
}

func (s *etcdClientSpy) Get(
	ctx context.Context,
	key string,
	opts ...clientv3.OpOption,
) (*clientv3.GetResponse, error) {
	return nil, nil
}

func (s *etcdClientSpy) Txn(ctx context.Context) clientv3.Txn {
	return s
}

func (s *etcdClientSpy) If(cs ...clientv3.Cmp) clientv3.Txn {
	s.ComparisonsSeenByIf = append(s.ComparisonsSeenByIf, cs...)

	return s
}

func (s *etcdClientSpy) Then(ops ...clientv3.Op) clientv3.Txn {
	if s.ShouldTxnIfSucceed {
		return s
	}

	s.OperationsSeenByThen = append(s.OperationsSeenByThen, ops...)

	return s
}

// Else is unimplemented as of now.
func (s *etcdClientSpy) Else(...clientv3.Op) clientv3.Txn {
	return s
}

func (s *etcdClientSpy) Commit() (*clientv3.TxnResponse, error) {
	return s.ReturnTxnResponse, s.CommitError
}

func TestSaveCalculation(t *testing.T) {
	cli := NewETCDClientSpy()

	calculation := store.Calculation{
		Name: "some-operation-name",
		Metadata: store.CalculationMetadata{
			Created: time.Now(),
		},
	}

	cli.ShouldTxnIfSucceed = true

	if err := calculation.Save(context.Background(), cli); err != nil {
		t.Errorf("unexpected error: %s", err)
	}

}
