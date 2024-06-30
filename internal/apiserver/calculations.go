package apiserver

import (
	"context"

	"github.com/vickleford/calculator/internal/pb"
)

var _ pb.CalculationsServer = &Calculations{}

type Calculations struct {
	pb.UnimplementedCalculationsServer
}

type operationsStore interface {
	Begin(context.Context) error
	Load(context.Context, string) error
	Save(context.Context) error
}

func NewCalculations() *Calculations {
	return &Calculations{}
}
