package worker

import (
	"context"
	"errors"
	"fmt"
	"log"

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
}

func NewFibOf(q queue, ds datastore) *FibOfWorker {
	return &FibOfWorker{queue: q, datastore: ds}
}

func (w *FibOfWorker) Start(ctx context.Context) error {
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

		// todo: set the operation's started at time.

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

		if err := w.datastore.Save(ctx, calculation); err != nil {
			log.Printf("error saving calculation %q: %s", job.OperationName, err)
			continue
			// TODO: If it has already been acked and I have failed to save the
			// status, how might I recover?
		}
	}
}
