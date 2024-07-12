package workqueue

import "fmt"

func NewAcknowledgementError(operation string, err, original error) *AcknowledgementError {
	return &AcknowledgementError{original: original, err: err, operation: operation}
}

type AcknowledgementError struct {
	original  error
	err       error
	operation string
}

func (e *AcknowledgementError) Error() string {
	msg := fmt.Sprintf("an acknowledgement error occured during %s: %s",
		e.operation, e.err)

	if e.original == nil {
		return msg
	}

	return fmt.Sprintf("while handling error %q, %s", e.original, msg)
}

func (e *AcknowledgementError) Unwrap() error {
	return e.err
}

const (
	AcknowledgementErrorOpReject = "reject"
	AcknowledgementErrorOpAck    = "ack"
	AcknowledgementErrorOpNack   = "nack"
)

func (e *AcknowledgementError) Operation() string {
	return e.operation
}
