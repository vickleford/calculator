package store_test

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
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
	client := store.NewCalculationStore(spy)

	calculation := store.Calculation{
		Name: "some-operation-name",
		Metadata: store.CalculationMetadata{
			Created: time.Now(),
		},
	}

	expectedKey := store.CalculationKey(calculation)

	if err := client.Save(context.Background(), calculation); err != nil {
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
	created := time.Now().Add(-10 * time.Second)
	started := time.Now()
	calculation := store.Calculation{
		Name: "some-operation-name",
		Metadata: store.CalculationMetadata{
			Created: created,
			Started: &started,
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
	client := store.NewCalculationStore(spy)

	spy.ShouldTxnIfSucceed = true

	if err := client.Save(context.Background(), calculation); err != nil {
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
		if actual.ValueBytes() == nil {
			t.Errorf("expected a nil value to have been written")
		}
		if value := actual.ValueBytes(); value != nil {
			savedCalculation := new(store.Calculation)
			if err := json.Unmarshal(value, savedCalculation); err != nil {
				t.Errorf("error unmarshaling value written to Calculation: %s", err)
			}
			if savedCalculation.Name != calculation.Name {
				t.Errorf("expected %q but got %q", calculation.Name, savedCalculation.Name)
			}
			if !created.Equal(savedCalculation.Metadata.Created) {
				t.Errorf("expected %q but got %q", created, savedCalculation.Metadata.Created)
			}
			if !started.Equal(*savedCalculation.Metadata.Started) {
				t.Errorf("expected %q but got %q", started, *savedCalculation.Metadata.Started)
			}
		}
	}
}

func TestCreateCalculation_WhenKeyDoesNotExist(t *testing.T) {
	calculation := store.Calculation{
		Name: "some-operation-name",
		Metadata: store.CalculationMetadata{
			Created: time.Now(),
		},
	}

	expectedKey := store.CalculationKey(calculation)

	spy := NewETCDClientSpy()
	spy.ShouldTxnIfSucceed = true
	spy.ReturnTxnResponse = &clientv3.TxnResponse{
		Succeeded: true,
	}

	client := store.NewCalculationStore(spy)

	if err := client.Create(context.Background(), calculation); err != nil {
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

func TestCreateCalculation_WhenKeyAlreadyExists(t *testing.T) {
	calculation := store.Calculation{
		Name: "some-operation-name",
		Metadata: store.CalculationMetadata{
			Created: time.Now(),
		},
	}

	expectedKey := store.CalculationKey(calculation)

	spy := NewETCDClientSpy()
	spy.ShouldTxnIfSucceed = false
	spy.ReturnTxnResponse = &clientv3.TxnResponse{
		Succeeded: false,
	}

	client := store.NewCalculationStore(spy)
	if err := client.Create(context.Background(), calculation); !errors.Is(err, store.ErrKeyAlreadyExists) {
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

	if len(spy.OperationsSeenByThen) != 0 {
		t.Errorf("should not have executed then but saw %d operations",
			len(spy.OperationsSeenByThen))
	}
}

func TestCalculationStore_Get_WhenKeyExists(t *testing.T) {
	calculation := store.Calculation{
		Name: "some-operation-name",
		Metadata: store.CalculationMetadata{
			Created: time.Now(),
		},
	}

	calculationMarshaled, err := json.Marshal(calculation)
	if err != nil {
		t.Fatalf("error setting up test with marshaled calculation: %s", err)
	}

	spy := NewETCDClientSpy()
	spy.ReturnGetResponse = &clientv3.GetResponse{
		Kvs: []*mvccpb.KeyValue{
			{Value: calculationMarshaled},
		},
		Count: 1,
	}

	client := store.NewCalculationStore(spy)
	actual, err := client.Get(context.Background(), calculation.Name)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	if actual.Name != calculation.Name {
		t.Errorf("got name %q but expected %q", actual.Name, calculation.Name)
	}

	if actual.Metadata.Started != calculation.Metadata.Started {
		t.Errorf("got started time %q but expected %q",
			actual.Metadata.Started, calculation.Metadata.Started)
	}
}

func TestCclaulationStore_Get_WhenKeyNotExists(t *testing.T) {
	spy := NewETCDClientSpy()
	spy.ReturnGetResponse = &clientv3.GetResponse{
		Kvs:   []*mvccpb.KeyValue{},
		Count: 0,
	}

	client := store.NewCalculationStore(spy)
	_, err := client.Get(context.Background(), "whatever")
	if !errors.Is(err, store.ErrKeyNotFound) {
		t.Errorf("unexpected error: %#v", err)
	}
}

func TestSetStarted(t *testing.T) {
	original := store.Calculation{
		Name: uuid.NewString(),
		Metadata: store.CalculationMetadata{
			Created: time.Now(),
		},
	}

	originalMarshaled, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("unable to set up test: %s", err)
	}

	var v int64 = 423893

	spy := NewETCDClientSpy()
	spy.ReturnGetResponse = &clientv3.GetResponse{
		Kvs: []*mvccpb.KeyValue{
			{
				Value:   originalMarshaled,
				Version: v,
			},
		},
		Count: 1,
	}

	spy.ShouldTxnIfSucceed = true
	spy.ReturnTxnResponse = &clientv3.TxnResponse{
		Succeeded: true,
	}

	started := time.Now()
	client := store.NewCalculationStore(spy)
	err = client.SetStartedTime(context.Background(), original.Name, started)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	expectedKey := store.CalculationKey(original)

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
		return
	}

	actual := spy.OperationsSeenByThen[0]
	if !actual.IsPut() {
		t.Errorf("expected a PUT operation")
	}
	if key := string(actual.KeyBytes()); key != expectedKey {
		t.Errorf("expected key %q but saw %q", expectedKey, key)
	}

	savedCalculation := store.Calculation{}
	if err := json.Unmarshal(actual.ValueBytes(), &savedCalculation); err != nil {
		t.Errorf("unable to unmarshal written result: %s", err)
	}

	if !started.Equal(*savedCalculation.Metadata.Started) {
		t.Errorf("expected time %q but got %q", started, *savedCalculation.Metadata.Started)
	}
}

func TestIntegration_CreateCalculation(t *testing.T) {
	etcdEndpoint := os.Getenv("ETCD_ENDPOINT")
	if etcdEndpoint == "" {
		t.Skip(`set ETCD_ENDPOINT to run this test, e.g. ETCD_ENDPOINT="localhost:2379"`)
	}

	calculation := store.Calculation{
		Name: uuid.New().String(),
		Metadata: store.CalculationMetadata{
			Created: time.Now(),
		},
	}

	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{etcdEndpoint},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		t.Fatalf("unable to set up client: %s", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	datastore := store.NewCalculationStore(cli)

	if err := datastore.Create(ctx, calculation); err != nil {
		t.Errorf("unable to create calculation: %s", err)
	}

	err = datastore.Create(ctx, calculation)
	if !errors.Is(err, store.ErrKeyAlreadyExists) {
		t.Errorf("expected key to already exist but got %#v", err)
	}

	key := store.CalculationKey(calculation)
	_, err = cli.Delete(ctx, key)
	if err != nil {
		t.Logf("error deleting key %q", key)
	}
}
