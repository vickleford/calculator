package workqueue

// FibonacciOfJob signals to begin a FibonacciOf calculation.
type FibonacciOfJob struct {
	// OperationName is the name of the operation in the data store.
	OperationName string `json:"operation_name"`
	// Start describes the starting number in the sequence.
	Start int64 `json:"start"`
	// Position describes which number in the sequence to calculate; it may be
	// considered synonomous with "index". The first number in the sequence is
	// position 1.
	Position int64 `json:"position"`
}
