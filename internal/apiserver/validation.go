package apiserver

import (
	"fmt"

	"cloud.google.com/go/longrunning/autogen/longrunningpb"
	"github.com/google/uuid"
)

func validateGetOperationRequest(r *longrunningpb.GetOperationRequest) error {
	_, err := uuid.Parse(r.Name)
	if err != nil {
		return fmt.Errorf("operation name must be a UUID")
	}

	return nil
}
