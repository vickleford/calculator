package apiserver

import (
	"context"
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
	store datastore
}

type datastore interface {
	Save(context.Context, store.Calculation) error
}

func NewCalculations(store datastore) *Calculations {
	return &Calculations{store: store}
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

	// Risk vs reward analysis would make some believe name collisions are too
	// close to impossible to worry about. In this context, it could give
	// another operation. We want to handle it at a later time, but can
	// procrastinate for now.
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

	if err := c.store.Save(ctx, calculation); err != nil {
		log.Printf("error saving calculation: %s", err)
		return nil, status.Error(codes.Internal, "internal error")
	}

	return op, nil
}
