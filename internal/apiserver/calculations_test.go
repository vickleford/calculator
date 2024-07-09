package apiserver_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/vickleford/calculator/internal/apiserver"
	"github.com/vickleford/calculator/internal/pb"
	"github.com/vickleford/calculator/internal/store"
)

type CreateFunc func(context.Context, store.Calculation) error

func (f CreateFunc) Create(ctx context.Context, c store.Calculation) error {
	return f(ctx, c)
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
	queue := &workQ{}

	var createCalled bool

	server := apiserver.NewCalculations(
		CreateFunc(func(ctx context.Context, c store.Calculation) error {
			createCalled = true
			return nil
		}),
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
