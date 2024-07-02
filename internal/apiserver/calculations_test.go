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

type SaveFunc func(context.Context, store.Calculation) error

func (f SaveFunc) Save(ctx context.Context, c store.Calculation) error {
	return f(ctx, c)
}

type workQ struct {
	message []byte
}

func (q *workQ) PublishJSON(ctx context.Context, msg []byte) error {
	if q.message == nil {
		q.message = make([]byte, len(msg))
	}

	if n := copy(q.message, msg); n != len(msg) {
		return fmt.Errorf("copied %d bytes but wanted %d", n, len(msg))
	}

	return nil
}

func TestFibonacciOf_Create(t *testing.T) {
	queue := &workQ{}

	var saveCalled bool

	server := apiserver.NewCalculations(
		SaveFunc(func(ctx context.Context, c store.Calculation) error {
			saveCalled = true
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

	if !saveCalled {
		t.Error("expected save to be called")
	}

	if _, err := uuid.Parse(response.Name); err != nil {
		t.Errorf("name is not a uuid or can't parse: %s", err)
	}

	calculation := store.Calculation{}
	if err := json.Unmarshal(queue.message, &calculation); err != nil {
		t.Errorf("published message was not a calculation")
	}

	if _, err := uuid.Parse(calculation.Name); err != nil {
		t.Errorf("expected a UUID name")
	}
}
