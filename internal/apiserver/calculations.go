package apiserver

import (
	"context"
	"errors"
	"fmt"
	"log"

	"cloud.google.com/go/longrunning/autogen/longrunningpb"
	"github.com/google/uuid"
	"github.com/vickleford/calculator/internal/pb"
	"github.com/vickleford/calculator/internal/store"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var _ pb.CalculationsServer = &Calculations{}

type Calculations struct {
	pb.UnimplementedCalculationsServer
	store      datastore
	fibOfWorkQ workqueue
}

type datastore interface {
	Create(context.Context, store.Calculation) error
	Get(context.Context, string) (store.Calculation, error)
}

type workqueue interface {
	PublishJSON(context.Context, any) error
}

func NewCalculations(store datastore, fibOfWorkQ workqueue) *Calculations {
	return &Calculations{
		store:      store,
		fibOfWorkQ: fibOfWorkQ,
	}
}

func (c *Calculations) FibonacciOf(
	ctx context.Context,
	req *pb.FibonacciOfRequest,
) (*longrunningpb.Operation, error) {
	metadata := &pb.CalculationMetadata{
		Created: timestamppb.Now(),
	}
	pbMeta, err := anypb.New(metadata)
	if err != nil {
		log.Printf("error creating calculation metadata: %s", err)
		return nil, status.Error(codes.Internal, "internal error")
	}

	op := &longrunningpb.Operation{
		Name:     uuid.New().String(),
		Metadata: pbMeta,
	}

	calculation := store.Calculation{
		Name: op.Name,
		Metadata: store.CalculationMetadata{
			Created: metadata.Created.AsTime(),
		},
	}

	// TODO: When it errors, it should generate a new name and try again. If it
	// still doesn't work, return an error.
	if err := c.store.Create(ctx, calculation); errors.Is(err, store.ErrKeyAlreadyExists) {
		log.Printf("tried to create calculation %s but it already exists", calculation.Name)
		return nil, status.Error(codes.AlreadyExists, "already exists")
	} else if err != nil {
		log.Printf("error saving calculation: %s", err)
		return nil, status.Error(codes.Internal, "internal error")
	}

	if err := c.fibOfWorkQ.PublishJSON(ctx, calculation); err != nil {
		// TODO: We need to additionally consider cleanup of the calculation in
		// the store, and furthermore what happens if that fails.
		log.Printf("error publishing to workQueue for %s: %s", calculation.Name, err)
		return nil, status.Error(codes.Internal, "internal error")
	}

	return op, nil
}

func (c *Calculations) GetOperation(
	ctx context.Context,
	req *longrunningpb.GetOperationRequest,
) (*longrunningpb.Operation, error) {
	calc, err := c.store.Get(ctx, req.Name)
	if errors.Is(err, store.ErrKeyNotFound) {
		return nil, status.Error(codes.NotFound,
			fmt.Sprintf("could not find operation %q", req.Name))
	} else if err != nil {
		log.Printf("error getting calculation: %s", err)
		return nil, status.Error(codes.Internal, "internal error")
	}

	op := &longrunningpb.Operation{
		Name: calc.Name,
		Done: calc.Done,
	}

	if calc.Error != nil {
		op.Result = &longrunningpb.Operation_Error{
			Error: calc.Error,
		}
	} else if calc.Result != nil {
		var resp *anypb.Any
		var err error

		switch v := calc.Result.(type) {
		case *pb.FibonacciOfResponse:
			resp, err = anypb.New(v)
		}
		if err != nil {
			log.Printf("error setting calculation result to Any: %s", err)
			return nil, status.Error(codes.Internal, "internal error")
		}

		op.Result = &longrunningpb.Operation_Response{
			Response: resp,
		}
	}

	return op, nil
}
