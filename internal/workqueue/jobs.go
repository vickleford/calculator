package workqueue

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
