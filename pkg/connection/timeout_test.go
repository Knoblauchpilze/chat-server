package connection

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnit_IsTimeoutError(t *testing.T) {
	type testCase struct {
		err             error
		expectedTimeout bool
	}

	testCases := []testCase{
		{
			err:             nil,
			expectedTimeout: false,
		},
		{
			err:             fmt.Errorf("some error"),
			expectedTimeout: false,
		},
		{
			err:             context.DeadlineExceeded,
			expectedTimeout: true,
		},
		{
			err:             fmt.Errorf("some error: %w", fmt.Errorf("another error")),
			expectedTimeout: false,
		},
		{
			err:             fmt.Errorf("%w", fmt.Errorf("another error")),
			expectedTimeout: false,
		},
		{
			err:             fmt.Errorf("failed: %w", context.DeadlineExceeded),
			expectedTimeout: true,
		},
		{
			err:             generateNestedError(context.DeadlineExceeded, 3),
			expectedTimeout: true,
		},
		{
			err:             generateNestedError(context.DeadlineExceeded, 6),
			expectedTimeout: false,
		},
	}

	for _, tc := range testCases {
		t.Run("", func(t *testing.T) {
			actual := isTimeoutError(tc.err)

			assert.Equal(t, tc.expectedTimeout, actual, "Error %v not recognized properly")
		})
	}
}

func generateNestedError(err error, depth int) error {
	out := err
	for i := 0; i < depth; i++ {
		out = fmt.Errorf("depth %d: %w", i, out)
	}
	return out
}
