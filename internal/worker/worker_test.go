package worker_test

import (
	"context"
	"testing"
	"time"

	"github.com/vickleford/calculator/internal/calculators"
	"github.com/vickleford/calculator/internal/pb"
	"github.com/vickleford/calculator/internal/store"
	"github.com/vickleford/calculator/internal/worker"
	"github.com/vickleford/calculator/internal/workqueue"
	"google.golang.org/grpc/codes"
)

type consumer struct {
	fibOfQueue    chan *workqueue.FibonacciOfJob
	fibOfQueueErr error
}

func (c *consumer) NextFibOfJob(context.Context) (*workqueue.FibonacciOfJob, error) {
	return <-c.fibOfQueue, c.fibOfQueueErr
}

type storeSpy struct {
	saveErr error
	saved   chan store.Calculation
}

func (s *storeSpy) Save(ctx context.Context, calc store.Calculation) error {
	s.saved <- calc
	return s.saveErr
}

func TestFibOfWorker_Successful(t *testing.T) {
	saved := make(chan store.Calculation)
	fakeStore := &storeSpy{saved: saved}

	queue := &consumer{fibOfQueue: make(chan *workqueue.FibonacciOfJob, 10)}
	job := &workqueue.FibonacciOfJob{
		OperationName: "george",
		First:         0,
		Second:        1,
		Position:      5,
	}
	queue.fibOfQueue <- job

	w := worker.NewFibOf(queue, fakeStore)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	go func() {
		err := w.Start(ctx)
		if err != nil {
			t.Errorf("unexpected error from worker start: %s", err)
		}
	}()

	var actual store.Calculation
	select {
	case msg := <-saved:
		actual = msg
	case <-ctx.Done():
		t.Fatalf("test timed out")
	}

	if !actual.Done {
		t.Error("expected Done to be set")
	}

	if actual.Error != nil {
		t.Errorf("unexpected error set: %#v", actual.Error)
	}

	if actual.Result == nil {
		t.Fatalf("expected result to be set but it was nil")
	}

	res, ok := actual.Result.(*pb.FibonacciOfResponse)
	if !ok {
		t.Fatalf("result was an unexpected type: %T", actual.Result)
	}

	if res.First != job.First {
		t.Errorf("expected %d but got %d", job.First, res.First)
	}

	if res.Second != job.Second {
		t.Errorf("expected %d but got %d", job.Second, res.Second)
	}

	if res.NthPosition != job.Position {
		t.Errorf("expected %d but got %d", job.Position, res.NthPosition)
	}

	if res.Result != 3 {
		t.Errorf("got wrong result: %d", res.Result)
	}
}

func TestFibOfWorker_Error(t *testing.T) {
	saved := make(chan store.Calculation)
	fakeStore := &storeSpy{saved: saved}

	queue := &consumer{fibOfQueue: make(chan *workqueue.FibonacciOfJob, 10)}
	job := &workqueue.FibonacciOfJob{
		OperationName: "george",
		First:         0,
		Second:        1,
		Position:      -5,
	}
	queue.fibOfQueue <- job

	w := worker.NewFibOf(queue, fakeStore)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	go func() {
		err := w.Start(ctx)
		if err != nil {
			t.Errorf("unexpected error from worker start: %s", err)
		}
	}()

	var actual store.Calculation
	select {
	case msg := <-saved:
		actual = msg
	case <-ctx.Done():
		t.Fatalf("test timed out")
	}

	if !actual.Done {
		t.Error("expected Done to be set")
	}

	if actual.Result != nil {
		t.Errorf("unexpected result set: %#v", actual.Result)
	}

	if actual.Error == nil {
		t.Fatalf("expected error to be set but it was nil")
	}

	if actual.Error.Code != int32(codes.InvalidArgument) {
		// Because the test sets position -5
		t.Errorf("expected invalid argument error code")
	}

	if actual.Error.Message != calculators.ErrFibonacciPositionInvalid.Error() {
		t.Errorf("unexpected error message: %q", actual.Error.Message)
	}
}
