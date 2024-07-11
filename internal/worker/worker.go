package worker

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/vickleford/calculator/internal/calculators"
	"github.com/vickleford/calculator/internal/pb"
	"github.com/vickleford/calculator/internal/store"
	"github.com/vickleford/calculator/internal/workqueue"
	"google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc/codes"
)

type FibOfWorker struct {
	queue     queue
	datastore datastore
}

type queue interface {
	NextFibOfJob(context.Context) (*workqueue.FibonacciOfJob, error)
}

type datastore interface {
	Save(context.Context, store.Calculation) error
	SetStartedTime(context.Context, string, time.Time) error
}

func NewFibOf(q queue, ds datastore) *FibOfWorker {
	return &FibOfWorker{queue: q, datastore: ds}
}

func (w *FibOfWorker) Start(ctx context.Context) error {
	// TODO: if the context ends, consider trying to write an aborted error to
	// the datastore prior to closing down.
	for {
		job, err := w.queue.NextFibOfJob(ctx)
		if err != nil {
			return fmt.Errorf("error getting next message: %w", err)
		}
		if job == nil {
			log.Println("warning: got nil job")
			continue
		}

		// ack? leave that to the workqueue consumer since acking is a rmq thing.

		// This is a weak area where a job could get lost. Give it a good
		// college effort.
		err = Retry(ctx, func() error {
			err := w.datastore.SetStartedTime(ctx, job.OperationName, time.Now())
			if err != nil {
				log.Printf("error setting job started time: %s", err)
			}
			return err
		})
		if err != nil {
			log.Printf("failed to set job started time on %q: %s",
				job.OperationName, err)
			continue
		} else {
			log.Printf("successfully set job started time for %q", job.OperationName)
		}

		c := calculators.NewFibonacci(job.First, job.Second)
		result, jobErr := c.NumberAtPosition(job.Position)

		calculation := store.Calculation{
			Name: job.OperationName,
			// TODO: Completion time.
			Done: true,
		}

		if jobErr != nil {
			log.Printf("error processing calculation %q: %s", job.OperationName, jobErr)

			state := &status.Status{
				Code:    int32(codes.Internal), // Default to internal.
				Message: jobErr.Error(),
			}

			if errors.Is(jobErr, calculators.ErrFibonacciPositionInvalid) {
				state.Code = int32(codes.InvalidArgument)
			}

			calculation.Error = state
		} else {
			calculation.Result = &pb.FibonacciOfResponse{
				First:       job.First,
				Second:      job.Second,
				NthPosition: job.Position,
				Result:      result,
			}
		}

		err = Retry(ctx, func() error {
			if err := w.datastore.Save(ctx, calculation); err != nil {
				log.Printf("error saving calculation %q: %s", job.OperationName, err)
				// TODO: If it has already been acked and I have failed to save the
				// status, how might I recover?
			}
			return err
		})
		if err != nil {
			log.Printf("failed to save calculation %q: %s", job.OperationName, err)
		} else {
			log.Printf("successfully saved calculation %q", job.OperationName)
		}
	}
}
