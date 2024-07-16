package worker_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/vickleford/calculator/internal/calculators"
	"github.com/vickleford/calculator/internal/store"
	"github.com/vickleford/calculator/internal/worker"
	"google.golang.org/grpc/codes"
)

type storeSpy struct {
	saveErr error
	saved   store.Calculation

	setStartedName string
	setStartedTime time.Time
	setStartedErr  error

	getFunc func(context.Context, string) (store.Calculation, error)
}

func (s *storeSpy) Get(ctx context.Context, name string) (store.Calculation, error) {
	if s.getFunc == nil {
		panic("unimplemented")
	}
	return s.getFunc(ctx, name)
}

func (s *storeSpy) SetStartedTime(ctx context.Context, name string, t time.Time) error {
	s.setStartedName = name
	s.setStartedTime = t
	return s.setStartedErr
}

func (s *storeSpy) Save(ctx context.Context, calc store.Calculation) error {
	s.saved = calc
	return s.saveErr
}

func FibonacciOfJobJSON(t *testing.T, j worker.FibonacciOfJob) []byte {
	t.Helper()
	b, err := json.Marshal(j)
	if err != nil {
		t.Errorf("error marshaling job to JSON: %s", err)
	}
	return b
}

func TestFibOfWorker_Successful(t *testing.T) {
	fakeStore := &storeSpy{}
	fakeStore.getFunc = func(context.Context, string) (store.Calculation, error) {
		return store.Calculation{
			Name: "george",
			Metadata: store.CalculationMetadata{
				Created: time.Now(),
				Version: 1,
			},
		}, nil
	}

	job := worker.FibonacciOfJob{
		OperationName: "george",
		First:         0,
		Second:        1,
		Position:      5,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	w := worker.NewFibOf(fakeStore)

	if err := w.Handle(ctx, FibonacciOfJobJSON(t, job)); err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	if !fakeStore.saved.Done {
		t.Error("expected Done to be set")
	}

	if fakeStore.saved.Error != nil {
		t.Errorf("unexpected error set: %#v", fakeStore.saved.Error)
	}

	if fakeStore.saved.Result == nil {
		t.Fatalf("expected result to be set but it was nil")
	}

	res := store.FibonacciOfResult{}
	if err := json.Unmarshal(fakeStore.saved.Result, &res); err != nil {
		t.Fatalf("unable to unmarshal result: %s", err)
	}

	if res.First != job.First {
		t.Errorf("expected %d but got %d", job.First, res.First)
	}

	if res.Second != job.Second {
		t.Errorf("expected %d but got %d", job.Second, res.Second)
	}

	if res.Position != job.Position {
		t.Errorf("expected %d but got %d", job.Position, res.Position)
	}

	if res.Result != 3 {
		t.Errorf("got wrong result: %d", res.Result)
	}
}

func TestFibOfWorker_Error(t *testing.T) {
	fakeStore := &storeSpy{}
	fakeStore.getFunc = func(context.Context, string) (store.Calculation, error) {
		return store.Calculation{
			Name: "george",
			Metadata: store.CalculationMetadata{
				Created: time.Now(),
				Version: 1,
			},
		}, nil
	}

	job := worker.FibonacciOfJob{
		OperationName: "george",
		First:         0,
		Second:        1,
		Position:      -5,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	w := worker.NewFibOf(fakeStore)
	if err := w.Handle(ctx, FibonacciOfJobJSON(t, job)); err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	if !fakeStore.saved.Done {
		t.Error("expected Done to be set")
	}

	if fakeStore.saved.Result != nil {
		t.Errorf("unexpected result set: %#v", fakeStore.saved.Result)
	}

	if fakeStore.saved.Error == nil {
		t.Fatalf("expected error to be set but it was nil")
	}

	if fakeStore.saved.Error.Code != int32(codes.InvalidArgument) {
		// Because the test sets position -5
		t.Errorf("expected invalid argument error code")
	}

	if fakeStore.saved.Error.Message != calculators.ErrFibonacciPositionInvalid.Error() {
		t.Errorf("unexpected error message: %q", fakeStore.saved.Error.Message)
	}
}

func TestFibOfWorker_SetsStartedTime(t *testing.T) {
	fakeStore := &storeSpy{}
	fakeStore.getFunc = func(context.Context, string) (store.Calculation, error) {
		return store.Calculation{
			Name: "george",
			Metadata: store.CalculationMetadata{
				Created: time.Now(),
				Version: 1,
			},
		}, nil
	}

	job := worker.FibonacciOfJob{
		OperationName: "george",
		First:         0,
		Second:        1,
		Position:      5,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	w := worker.NewFibOf(fakeStore)
	if err := w.Handle(ctx, FibonacciOfJobJSON(t, job)); err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	if fakeStore.setStartedTime.IsZero() {
		t.Error("the started time was not set")
	} else if time.Since(fakeStore.setStartedTime) > 5*time.Second {
		t.Errorf("the started time %q is older than expected",
			fakeStore.setStartedTime)
	}

	if now := time.Now(); fakeStore.setStartedTime.After(now) {
		t.Errorf("the started time %q is after now %q",
			fakeStore.setStartedTime, now)
	}

	if fakeStore.setStartedName != job.OperationName {
		t.Errorf("got %q but expected %q", fakeStore.setStartedName, job.OperationName)
	}
}
