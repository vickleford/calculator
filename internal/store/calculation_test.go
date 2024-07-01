package store_test

import (
	"context"
	"testing"
	"time"

	"github.com/vickleford/calculator/internal/store"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type etcdClientSpy struct {
	// ComparisonsSeenByIf are the comparisons observed by If.
	ComparisonsSeenByIf []clientv3.Cmp

	// OperationsSeenByThen are the operations seen by Then.
	OperationsSeenByThen []clientv3.Op

	// WritesSeenByPut records keys and values of Put, but not if it is in a
	// transaction via OpPut.
	WritesSeenByPut map[string]string

	// PutError is the error returned by Put
	PutError error

	// ReturnGetResponse is what the spy will return when Get is called.
	ReturnGetResponse *clientv3.GetResponse
	// GetError is the error returned by Get.
	GetError error

	// ReturnTxnResponse is the transaction response to return from Commit.
	ReturnTxnResponse *clientv3.TxnResponse

	// CommitError is the error returned by Commit.
	CommitError error

	// ShouldTxnIfSucceed tells the spy whether the If from a Txn should match or
	// not. If not (false), it will not execute the operations inside If.
	ShouldTxnIfSucceed bool
}

func NewETCDClientSpy() *etcdClientSpy {
	return &etcdClientSpy{
		ComparisonsSeenByIf:  make([]clientv3.Cmp, 0),
		OperationsSeenByThen: make([]clientv3.Op, 0),
		WritesSeenByPut:      make(map[string]string),
	}
}

func (s *etcdClientSpy) Get(
	ctx context.Context,
	key string,
	opts ...clientv3.OpOption,
) (*clientv3.GetResponse, error) {
	return s.ReturnGetResponse, s.GetError
}

func (s *etcdClientSpy) Put(
	ctx context.Context,
	key, value string,
	opts ...clientv3.OpOption,
) (*clientv3.PutResponse, error) {
	s.WritesSeenByPut[key] = value
	return nil, s.PutError
}

func (s *etcdClientSpy) Txn(ctx context.Context) clientv3.Txn {
	return s
}

func (s *etcdClientSpy) If(cs ...clientv3.Cmp) clientv3.Txn {
	s.ComparisonsSeenByIf = append(s.ComparisonsSeenByIf, cs...)

	return s
}

func (s *etcdClientSpy) Then(ops ...clientv3.Op) clientv3.Txn {
	if !s.ShouldTxnIfSucceed {
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

func TestSaveCalculation_WhenKeyDoesNotExist(t *testing.T) {
	spy := NewETCDClientSpy()
	spy.ReturnGetResponse = &clientv3.GetResponse{Count: 0}
	client := store.NewEtcdClient(spy)

	calculation := store.Calculation{
		Name: "some-operation-name",
		Metadata: store.CalculationMetadata{
			Created: time.Now(),
		},
	}

	expectedKey := store.CalculationKey(calculation)

	if err := client.SaveCalculation(context.Background(), calculation); err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	t.Logf("puts: %#v", spy.WritesSeenByPut)

	if len(spy.WritesSeenByPut) != 1 {
		t.Fatalf("got %d writes to put", len(spy.WritesSeenByPut))
	}

	if _, ok := spy.WritesSeenByPut[expectedKey]; !ok {
		t.Errorf("did not see a write to key %q", expectedKey)
	}
}

func TestSaveCalculation_WhenKeyExists(t *testing.T) {
	calculation := store.Calculation{
		Name: "some-operation-name",
		Metadata: store.CalculationMetadata{
			Created: time.Now(),
		},
	}

	expectedKey := store.CalculationKey(calculation)

	spy := NewETCDClientSpy()
	spy.ReturnGetResponse = &clientv3.GetResponse{
		Count: 1,
		Kvs: []*mvccpb.KeyValue{
			{
				Key:     []byte(expectedKey),
				Version: 428922,
			},
		},
	}
	spy.ReturnTxnResponse = &clientv3.TxnResponse{
		Succeeded: true,
	}
	client := store.NewEtcdClient(spy)

	spy.ShouldTxnIfSucceed = true

	if err := client.SaveCalculation(context.Background(), calculation); err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	if len(spy.ComparisonsSeenByIf) != 1 {
		t.Errorf("unexpected if comparisons: %#v", spy.ComparisonsSeenByIf)
	} else {
		actual := spy.ComparisonsSeenByIf[0]
		if string(actual.Key) != expectedKey {
			t.Errorf("saw key %q but expected %q", actual.Key, expectedKey)
		}
	}

	if len(spy.OperationsSeenByThen) != 1 {
		t.Errorf("saw %d operations", len(spy.OperationsSeenByThen))
	} else {
		actual := spy.OperationsSeenByThen[0]
		if !actual.IsPut() {
			t.Errorf("expected a PUT operation")
		}
		if key := string(actual.KeyBytes()); key != expectedKey {
			t.Errorf("expected key %q but saw %q", expectedKey, key)
		}
	}
}
