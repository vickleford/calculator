package worker

import (
	"context"
	"time"
)

type Task func() error

func Retry(ctx context.Context, f Task) error {
	var delay, ceiling, factor time.Duration = 30 * time.Millisecond, 10 * time.Minute, 2

	for {
		if err := f(); err == nil {
			return nil
		}

		wait := time.NewTicker(delay)

		select {
		case <-wait.C:
			delay = delay * factor
			if delay > ceiling {
				delay = ceiling
			}
		case <-ctx.Done():
		}

		wait.Stop()

		if ctx.Err() != nil {
			return ctx.Err()
		}
	}
}
