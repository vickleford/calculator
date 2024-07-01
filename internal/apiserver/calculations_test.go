package apiserver_test

import (
	"context"
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

func TestFibonacciOf_Create(t *testing.T) {
	var saveCalled bool

	server := apiserver.NewCalculations(SaveFunc(func(ctx context.Context, c store.Calculation) error {
		saveCalled = true
		return nil
	}))

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
}
