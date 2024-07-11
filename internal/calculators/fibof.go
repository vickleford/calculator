package calculators

import "errors"

var ErrFibonacciPositionInvalid = errors.New("Fibonacci number sequences start at position 1")

type Fibonacci struct {
	// first defines the first number in the sequence at position 1.
	first int64
	// second defines the second number in the sequence at position 2.
	second int64
}

func NewFibonacci(first, second int64) *Fibonacci {
	return &Fibonacci{first: first, second: second}
}

func (f *Fibonacci) NumberAtPosition(position int64) (int64, error) {
	if position < 1 {
		return -1, ErrFibonacciPositionInvalid
	}

	if position == 1 {
		return f.first, nil
	}
	if position == 2 {
		return f.second, nil
	}

	var twoBefore, previous = f.first, f.second
	var result int64
	for i := int64(3); i <= position; i++ {
		result = twoBefore + previous
		twoBefore, previous = previous, result
	}

	return result, nil
}
