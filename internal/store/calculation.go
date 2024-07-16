package store

import (
	"encoding/json"
	"time"

	"google.golang.org/genproto/googleapis/rpc/status"
)

type CalculationMetadata struct {
	Created time.Time  `json:"created"`
	Started *time.Time `json:"started,omitempty"`

	// Version carries the version identifier stored of the Calculation.
	Version int64 `json:"-"`
}

type CalculationError struct {
	Message string   `json:"message"`
	Details []string `json:"details"` // todo: revisit this.
}

type Calculation struct {
	Name     string              `json:"name"`
	Metadata CalculationMetadata `json:"metadata"`
	Done     bool                `json:"done"`

	// Error is mutually exclusive with Response. It should only be set whe done
	// is true.
	Error *status.Status `json:"error,omitempty"`
	// Response is mutually exclusive with Error. It should only be set whe done
	// is true.
	Result json.RawMessage `json:"result,omitempty"` // todo: could this instead use generics?
}

type FibonacciOfResult struct {
	Position int64 `json:"position"`
	First    int64 `json:"first"`
	Second   int64 `json:"second"`
	Result   int64 `json:"result"`
}
