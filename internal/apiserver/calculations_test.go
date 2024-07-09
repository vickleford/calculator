package apiserver_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"cloud.google.com/go/longrunning/autogen/longrunningpb"
	"github.com/google/uuid"
	"github.com/vickleford/calculator/internal/apiserver"
	"github.com/vickleford/calculator/internal/pb"
	"github.com/vickleford/calculator/internal/store"
)

type fakeStore struct {
	CreateFunc func(context.Context, store.Calculation) error
	GetFunc    func(context.Context, string) (store.Calculation, error)
}

func (s fakeStore) Create(ctx context.Context, c store.Calculation) error {
	if s.CreateFunc == nil {
		panic("Create is unimplemented")
	}

	return s.CreateFunc(ctx, c)
}

func (s fakeStore) Get(ctx context.Context, key string) (store.Calculation, error) {
	if s.GetFunc == nil {
		panic("Get is unimplemented")
	}

	return s.GetFunc(ctx, key)
}

type workQ struct {
	message []byte
}

func (q *workQ) PublishJSON(ctx context.Context, msg any) error {
	b, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("error marshaling message: %w", err)
	}

	if q.message == nil {
		q.message = make([]byte, len(b))
	}

	if n := copy(q.message, b); n != len(b) {
		return fmt.Errorf("copied %d bytes but wanted %d", n, len(b))
	}

	return nil
}

func TestFibonacciOf_Create(t *testing.T) {
	var createCalled bool

	queue := &workQ{}
	mockStore := fakeStore{
		CreateFunc: func(ctx context.Context, c store.Calculation) error {
			createCalled = true
			return nil
		},
	}

	server := apiserver.NewCalculations(
		mockStore,
		queue,
	)

	ctx := context.Background()

	req := &pb.FibonacciOfRequest{Start: 1, NthPosition: 5}
	response, err := server.FibonacciOf(ctx, req)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	if !createCalled {
		t.Error("expected Create to be called")
	}

	storeUUID, err := uuid.Parse(response.Name)
	if err != nil {
		t.Errorf("name is not a uuid or can't parse: %s", err)
	}

	calculation := store.Calculation{}
	if err := json.Unmarshal(queue.message, &calculation); err != nil {
		t.Errorf("published message was not a calculation")
	}

	msgUUID, err := uuid.Parse(calculation.Name)
	if err != nil {
		t.Errorf("expected a UUID name")
	}

	if storeUUID != msgUUID {
		t.Errorf("message UUID %q and store UUID %q must be the same",
			msgUUID, storeUUID)
	}
}

func TestCalculations_GetOperation(t *testing.T) {
	createdAt := time.Now().Add(-30 * time.Second)
	opName := uuid.New().String()

	tests := []struct {
		Name    string
		GetFunc func(context.Context, string) (store.Calculation, error)
		assert  func(*testing.T, *longrunningpb.Operation, error)
	}{
		{
			Name: "OperationNotDone",
			GetFunc: func(ctx context.Context, key string) (store.Calculation, error) {
				return store.Calculation{
					Name: key,
					Metadata: store.CalculationMetadata{
						Created: createdAt,
					},
				}, nil
			},
			assert: func(t *testing.T, op *longrunningpb.Operation, err error) {
				if err != nil {
					t.Errorf("unexpected error: %s", err)
				}

				if op.Name != opName {
					t.Errorf("expected name %q but got %q", opName, op.Name)
				}

				if op.Done {
					t.Errorf("expected Done to be false")
				}
			},
		},
	}

	for _, test := range tests {
		test := test

		t.Run(test.Name, func(t *testing.T) {
			mockStore := fakeStore{
				GetFunc: test.GetFunc,
			}

			server := apiserver.NewCalculations(mockStore, nil)

			req := &longrunningpb.GetOperationRequest{Name: opName}

			response, err := server.GetOperation(context.Background(), req)

			test.assert(t, response, err)
		})
	}
}
