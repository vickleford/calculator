package apiserver

import (
	"context"
	"encoding/json"
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
	Save(context.Context, store.Calculation) error
}

type workqueue interface {
	PublishJSON(context.Context, []byte) error
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
	calculationJSON, err := json.Marshal(calculation)
	if err != nil {
		log.Printf("error marshaling calculation to JSON: %s", err)
		return nil, status.Error(codes.Internal, "internal error")
	}

	// TODO: this is incorrect. It shouldn't create or update. It should create
	// or error. When it errors, it should generate a new name and try again. If
	// it still doesn't work, return an error.
	if err := c.store.Save(ctx, calculation); err != nil {
		log.Printf("error saving calculation: %s", err)
		return nil, status.Error(codes.Internal, "internal error")
	}

	if err := c.fibOfWorkQ.PublishJSON(ctx, calculationJSON); err != nil {
		// TODO: We need to additionally consider cleanup of the calculation in
		// the store, and furthermore what happens if that fails.
		log.Printf("error creating workQueue for %s: %s", calculation.Name, err)
		return nil, status.Error(codes.Internal, "internal error")
	}

	return op, nil
}
