package worker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/vickleford/calculator/internal/calculators"
	"github.com/vickleford/calculator/internal/store"
	"google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc/codes"
)

// FibonacciOfJob signals to begin a FibonacciOf calculation.
type FibonacciOfJob struct {
	// OperationName is the name of the operation in the data store.
	OperationName string `json:"operation_name"`
	// First describes the first number in the sequence.
	First int64 `json:"first"`
	// Second describes the second number in the sequence.
	Second int64 `json:"second"`
	// Position describes which number in the sequence to calculate; it may be
	// considered synonomous with "index". The first number in the sequence is
	// position 1.
	Position int64 `json:"position"`
}

type FibOfHandler struct {
	datastore datastore
}

type datastore interface {
	Get(context.Context, string) (store.Calculation, error)
	Save(context.Context, store.Calculation) error
	SetStartedTime(context.Context, string, time.Time) error
}

func NewFibOf(ds datastore) *FibOfHandler {
	return &FibOfHandler{datastore: ds}
}

func (w *FibOfHandler) Handle(ctx context.Context, payload []byte) error {
	var job FibonacciOfJob
	if err := json.Unmarshal(payload, &job); err != nil {
		return fmt.Errorf("unable to unmarshal payload: %w", err)
	}

	// TODO: Consider a validate function on the FibonacciOfJob instead.
	if job.OperationName == "" {
		return fmt.Errorf("job has no operation name; payload: %s", payload)
	}

	// This is a weak area where a job could get lost. Give it a good
	// college effort.
	err := Retry(ctx, func() error {
		err := w.datastore.SetStartedTime(ctx, job.OperationName, time.Now())
		if err != nil {
			err = fmt.Errorf("error setting job started time on %q: %w", job.OperationName, err)
			log.Println(err)
			return err
		}
		return nil
	})
	if err != nil {
		return err
	} else {
		log.Printf("successfully set job started time for %q", job.OperationName)
	}

	c := calculators.NewFibonacci(job.First, job.Second)
	solution, jobErr := c.NumberAtPosition(job.Position)

	calculation, err := w.datastore.Get(ctx, job.OperationName)
	if err != nil {
		return fmt.Errorf("error getting calculation from store: %w", err)
	}

	calculation.Done = true
	// TODO: add job completed time.

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
		result := store.FibonacciOfResult{
			First:    job.First,
			Second:   job.Second,
			Position: job.Position,
			Result:   solution,
		}
		b, err := json.Marshal(result)
		if err != nil {
			return fmt.Errorf("error marshaling result: %w", err)
		}
		calculation.Result = b
	}

	err = Retry(ctx, func() error {
		if err := w.datastore.Save(ctx, calculation); err != nil {
			err = fmt.Errorf("error saving calculation %q: %w", job.OperationName, err)
			log.Println(err)
			return err
		}
		return nil
	})
	if err != nil {
		return err
	} else {
		log.Printf("successfully saved calculation %q", job.OperationName)
	}

	return nil
}
