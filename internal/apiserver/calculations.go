package apiserver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"cloud.google.com/go/longrunning/autogen/longrunningpb"
	"github.com/google/uuid"
	"github.com/vickleford/calculator/internal/pb"
	"github.com/vickleford/calculator/internal/store"
	"github.com/vickleford/calculator/internal/worker"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var _ pb.CalculationsServer = &Calculations{}

type Calculations struct {
	pb.UnimplementedCalculationsServer
	store      datastore
	fibOfWorkQ queue
}

type datastore interface {
	Create(context.Context, store.Calculation) error
	Get(context.Context, string) (store.Calculation, error)
}

type queue interface {
	PublishJSON(context.Context, any) error
}

func NewCalculations(store datastore, fibOfWorkQ queue) *Calculations {
	return &Calculations{
		store:      store,
		fibOfWorkQ: fibOfWorkQ,
	}
}

func (c *Calculations) FibonacciOf(
	ctx context.Context,
	req *pb.FibonacciOfRequest,
) (*longrunningpb.Operation, error) {
	// A typical implementation should validate the request. For example,
	// requesting the Fibonacci at positions 0 or -5 should not be allowed. This
	// implementation foregoes validation in order to exercise jobs failing
	// since it is more difficult to get this simple thing to fail with long
	// executions.

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
	// TODO: We need to additionally consider cleanup of the calculation in
	// the store, and furthermore what happens if that fails.
	if err := c.store.Create(ctx, calculation); errors.Is(err, store.ErrKeyAlreadyExists) {
		log.Printf("tried to create calculation %s but it already exists", calculation.Name)
		return nil, status.Error(codes.AlreadyExists, "already exists")
	} else if err != nil {
		log.Printf("error saving calculation: %s", err)
		return nil, status.Error(codes.Internal, "internal error")
	}

	job := worker.FibonacciOfJob{
		OperationName: calculation.Name,
		First:         req.First,
		Second:        req.Second,
		Position:      req.NthPosition,
	}

	if err := c.fibOfWorkQ.PublishJSON(ctx, job); err != nil {
		log.Printf("error publishing to workQueue for %s: %s", calculation.Name, err)
		return nil, status.Error(codes.Internal, "internal error")
	}

	return op, nil
}

func (c *Calculations) GetOperation(
	ctx context.Context,
	req *longrunningpb.GetOperationRequest,
) (*longrunningpb.Operation, error) {
	if err := validateGetOperationRequest(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	calc, err := c.store.Get(ctx, req.Name)
	if errors.Is(err, store.ErrKeyNotFound) {
		return nil, status.Error(codes.NotFound,
			fmt.Sprintf("could not find operation %q", req.Name))
	} else if err != nil {
		log.Printf("error getting calculation: %s", err)
		return nil, status.Error(codes.Internal, "internal error")
	}

	metadata := &pb.CalculationMetadata{
		Created: timestamppb.New(calc.Metadata.Created),
	}
	if calc.Metadata.Started != nil {
		metadata.Started = timestamppb.New(*calc.Metadata.Started)
	}
	metadataAsAnyPB, err := anypb.New(metadata)
	if err != nil {
		log.Printf("error marshaling calculation %q metadata to proto: %s", calc.Name, err)
	}

	op := &longrunningpb.Operation{
		Name:     calc.Name,
		Metadata: metadataAsAnyPB,
		Done:     calc.Done,
	}

	if calc.Error != nil {
		op.Result = &longrunningpb.Operation_Error{
			Error: calc.Error,
		}
	} else if calc.Result != nil {
		var resp *anypb.Any
		var err error

		result := store.FibonacciOfResult{}
		if err := json.Unmarshal(calc.Result, &result); err != nil {
			log.Printf("error unmarshaling calculation result: %s", err)
			return nil, status.Error(codes.Internal, "unsupported calculation result type")
		}

		fibonacciOfResponse := &pb.FibonacciOfResponse{
			First:       result.First,
			Second:      result.Second,
			NthPosition: result.Position,
			Result:      result.Result,
		}

		resp, err = anypb.New(fibonacciOfResponse)
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
