package apiserver_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"cloud.google.com/go/longrunning/autogen/longrunningpb"
	"github.com/google/uuid"
	"github.com/vickleford/calculator/internal/apiserver"
	"github.com/vickleford/calculator/internal/pb"
	"github.com/vickleford/calculator/internal/store"
	"github.com/vickleford/calculator/internal/worker"
	"google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc/codes"
	grpc_status "google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"
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

	req := &pb.FibonacciOfRequest{First: 1, Second: 1, NthPosition: 5}
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

	fibOfJob := worker.FibonacciOfJob{}
	if err := json.Unmarshal(queue.message, &fibOfJob); err != nil {
		t.Errorf("error unmarshaling job: %s", err)
	}

	msgUUID, err := uuid.Parse(fibOfJob.OperationName)
	if err != nil {
		t.Errorf("expected a UUID name")
	}

	if fibOfJob.First != req.First {
		t.Errorf("expected job to have first of %d but was %d",
			req.First, fibOfJob.First)
	}

	if fibOfJob.Second != req.Second {
		t.Errorf("expected job to have second of %d but was %d",
			req.Second, fibOfJob.Second)
	}

	if fibOfJob.Position != req.NthPosition {
		t.Errorf("expected job to have position of %d but was %d",
			req.NthPosition, fibOfJob.Position)
	}

	if storeUUID != msgUUID {
		t.Errorf("message UUID %q and store UUID %q must be the same",
			msgUUID, storeUUID)
	}
}

func TestCalculations_GetOperation(t *testing.T) {
	createdAt := time.Now().Add(-30 * time.Second)
	startedAt := time.Now().Add(-25 * time.Second)
	opName := uuid.New().String()

	fibOfResponse := store.FibonacciOfResult{
		First:    0,
		Second:   1,
		Position: 6,
		Result:   5,
	}
	fibOfResponseBytes, err := json.Marshal(fibOfResponse)
	if err != nil {
		t.Fatalf("unable to set up test with FibOfResponse raw json: %s", err)
	}

	retry := &errdetails.RetryInfo{RetryDelay: &durationpb.Duration{Seconds: 15}}
	retryAsAny, err := anypb.New(retry)
	if err != nil {
		t.Fatalf("error setting up test with anypb.Any for retry info: %s", err)
	}

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
		{
			Name: "OperationInError",
			GetFunc: func(ctx context.Context, key string) (store.Calculation, error) {
				return store.Calculation{
					Name: key,
					Metadata: store.CalculationMetadata{
						Created: createdAt,
						Started: &startedAt,
					},
					Done: true,
					Error: &status.Status{
						Code:    int32(code.Code_ABORTED),
						Message: "it all happened so fast",
						Details: []*anypb.Any{retryAsAny},
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

				if !op.Done {
					t.Error("operations with error should be marked done")
				}

				if actual := op.GetError().Code; actual != int32(code.Code_ABORTED) {
					t.Errorf("expected code %d but got %d", code.Code_ABORTED, actual)
				}
				if actual := op.GetError().Message; actual != "it all happened so fast" {
					t.Errorf("got wrong message: %q", actual)
				}
				if n := len(op.GetError().Details); n != 1 {
					t.Errorf("expected 1 error detail to be set but saw %d", n)
				} else {
					actual := &errdetails.RetryInfo{}
					if err := op.GetError().Details[0].UnmarshalTo(actual); err != nil {
						t.Errorf("unable to unmarshal protobuf message: %q", err)
					}
					if actual.RetryDelay.Seconds != retry.RetryDelay.Seconds {
						t.Errorf("expected %d retry delay seconds but got %d",
							retry.RetryDelay.Seconds,
							actual.RetryDelay.Seconds)
					}
				}
			},
		},
		{
			Name: "OperationCompletedYieldsResult",
			GetFunc: func(ctx context.Context, key string) (store.Calculation, error) {
				return store.Calculation{
					Name: key,
					Metadata: store.CalculationMetadata{
						Created: createdAt,
						Started: &startedAt,
					},
					Done:   true,
					Result: fibOfResponseBytes,
				}, nil
			},
			assert: func(t *testing.T, op *longrunningpb.Operation, err error) {
				if err != nil {
					t.Errorf("unexpected error: %s", err)
				}

				if op.Name != opName {
					t.Errorf("expected name %q but got %q", opName, op.Name)
				}

				if !op.Done {
					t.Errorf("operations with error should be marked done")
				}

				if actual := op.GetError(); actual != nil {
					t.Errorf("expected no error but got %#v", actual)
				}

				if op.GetResponse() == nil {
					t.Error("expected response to be set")
					return
				}

				actual := &pb.FibonacciOfResponse{}
				if err := op.GetResponse().UnmarshalTo(actual); err != nil {
					t.Errorf("unexpected error converting to desired pb type: %s", err)
				}

				if actual.First != fibOfResponse.First {
					t.Errorf("expected first to be %d but was %d",
						fibOfResponse.First,
						actual.First)
				}
				if actual.Second != fibOfResponse.Second {
					t.Errorf("expected second to be %d but was %d",
						fibOfResponse.Second,
						actual.Second)
				}
				if actual.NthPosition != fibOfResponse.Position {
					t.Errorf("expected nth position to be %d but was %d",
						fibOfResponse.Position,
						actual.NthPosition)
				}
				if actual.Result != fibOfResponse.Result {
					t.Errorf("expected result to be %d but was %d",
						fibOfResponse.Result,
						actual.Result)
				}
			},
		},
		{
			Name: "CalculationNotFound",
			GetFunc: func(ctx context.Context, key string) (store.Calculation, error) {
				return store.Calculation{}, store.ErrKeyNotFound
			},
			assert: func(t *testing.T, op *longrunningpb.Operation, err error) {
				if op != nil {
					t.Errorf("expected a nil operation but got %#v", op)
				}

				statusErr, ok := grpc_status.FromError(err)
				if !ok {
					t.Errorf("expected a gRPC status error")
				}

				if statusErr.Code() != codes.NotFound {
					t.Errorf("unexpected code: %s", statusErr.Code())
				}

				expected := fmt.Sprintf("could not find operation %q", opName)
				if actual := statusErr.Message(); actual != expected {
					t.Errorf("unexpected message: %q", actual)
				}
			},
		},
		{
			Name: "ErrorGettingCalculation",
			GetFunc: func(ctx context.Context, key string) (store.Calculation, error) {
				return store.Calculation{}, errors.New("oh no")
			},
			assert: func(t *testing.T, op *longrunningpb.Operation, err error) {
				if op != nil {
					t.Errorf("expected a nil operation but got %#v", op)
				}

				statusErr, ok := grpc_status.FromError(err)
				if !ok {
					t.Errorf("expected a gRPC status error")
				}

				if statusErr.Code() != codes.Internal {
					t.Errorf("unexpected code: %s", statusErr.Code())
				}

				if actual := statusErr.Message(); actual != "internal error" {
					t.Errorf("unexpected message: %q", actual)
				}
			},
		},
	}

	for _, test := range tests {
		test := test

		t.Run(test.Name, func(t *testing.T) {
			t.Parallel()

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
