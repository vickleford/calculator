package worker_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/vickleford/calculator/internal/worker"
)

func TestRetry_WhenFirstTrySucceeds(t *testing.T) {
	var executions int

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := worker.Retry(ctx, func() error {
		executions++
		return nil
	})
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	if executions != 1 {
		t.Errorf("saw %d executions", executions)
	}
}

func TestRetry_WhenContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	var executions int

	err := worker.Retry(ctx, func() error {
		executions++
		return fmt.Errorf("keep going")
	})
	if !errors.Is(err, context.Canceled) {
		t.Errorf("unexpected error: %#v", err)
	}
}

func TestRetry_WhenContextDeadlineExceeded(t *testing.T) {
	// Not terribly happy about how this duration is coupled to the
	// implementation but I don't have a real need to configure it yet.
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Millisecond)
	defer cancel()

	var executions int

	err := worker.Retry(ctx, func() error {
		executions++
		return fmt.Errorf("keep going")
	})
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("unexpected error: %#v", err)
	}

	if executions != 2 {
		t.Errorf("got %d executions", executions)
	}
}

func TestRetry_WhenItEventuallySucceeds(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	executions, failureBudget := 0, 2

	err := worker.Retry(ctx, func() error {
		executions++
		if failureBudget == 0 {
			return nil
		}
		failureBudget--
		return fmt.Errorf("keep going")
	})
	if err != nil {
		t.Errorf("unexpected error: %#v", err)
	}

	// expected executions = 1 + failureBudget's initial value.
	if executions != 3 {
		t.Errorf("got %d executions", executions)
	}
}
