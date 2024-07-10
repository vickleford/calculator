package calculators_test

import (
	"errors"
	"testing"

	"github.com/vickleford/calculator/internal/calculators"
)

func TestFibonacci_NumberAtPosition(t *testing.T) {
	// Given the sequence 0, 1, 1, 2, 3, 5:
	// 0: position 1, first
	// 1: position 2, second
	// 1: position 3
	// 2: position 4
	// 3: position 5
	// 5: position 6

	tests := []struct {
		name     string
		first    int64
		second   int64
		position int64
		expected int64

		// errorAssertion defines a custom assertion to make on the returned
		// error. When nil, the test will assert the error is also non-nil.
		errorAssertion func(*testing.T, error)
	}{
		{
			// 0, 1, 1
			name:     "StartAt0And1ThenFindThirdNumber",
			first:    0,
			second:   1,
			position: 3,
			expected: 1,
		},
		{
			// 0, 1, 1, 2, 3, 5
			name:     "StartAt0And1ThenFindTheSixthNumber",
			first:    0,
			second:   1,
			position: 6,
			expected: 5,
		},
		{
			name:     "NegativePositionIsAnError",
			first:    0,
			second:   1,
			position: -1,
			expected: -1,
			errorAssertion: func(t *testing.T, err error) {
				if !errors.Is(err, calculators.ErrFibonacciPositionInvalid) {
					t.Errorf("unexpected error: %#v", err)
				}
			},
		},
		{
			name:     "PositionZeroIsAnError",
			first:    0,
			second:   1,
			position: 0,
			expected: -1,
			errorAssertion: func(t *testing.T, err error) {
				if !errors.Is(err, calculators.ErrFibonacciPositionInvalid) {
					t.Errorf("unexpected error: %#v", err)
				}
			},
		},
		{
			name:     "FirstPositionGivesFirstNumber",
			first:    32,
			second:   41,
			position: 1,
			expected: 32,
		},
		{
			name:     "SecondPositionGivesSecondNumber",
			first:    0,
			second:   1,
			position: 2,
			expected: 1,
		},
		{
			// -50, 8, -42, -34, -76, -110
			name:     "NegativeFirstNumber",
			first:    -50,
			second:   8,
			position: 6,
			expected: -110,
		},
		{
			// 60, -23, 37, 14, 51
			name:     "NegativeSecondNumber",
			first:    60,
			second:   -23,
			position: 5,
			expected: 51,
		},
		{
			// -1, -2, -3, -5, -8
			name:     "NegativeNumbersAtPositionsOneAndTwo",
			first:    -1,
			second:   -2,
			position: 5,
			expected: -8,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			f := calculators.NewFibonacci(test.first, test.second)
			actual, err := f.NumberAtPosition(test.position)
			if test.errorAssertion != nil {
				test.errorAssertion(t, err)
			} else if err != nil {
				t.Errorf("unexpected error: %#v", err)
			}

			if actual != test.expected {
				t.Errorf("expected %d but got %d", test.expected, actual)
			}
		})
	}
}
