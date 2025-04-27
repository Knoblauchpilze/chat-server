package connection

import (
	"context"
	"errors"
)

func isTimeoutError(err error) bool {
	const maxNestedErrorDepth = 5
	return isTimeoutErrorRecursive(err, 0, maxNestedErrorDepth)
}

func isTimeoutErrorRecursive(err error, depth int, maxDepth int) bool {
	if err == nil {
		return false
	}

	if err == context.DeadlineExceeded {
		return true
	}

	cause := errors.Unwrap(err)
	if cause == nil {
		return false
	}

	if depth >= maxDepth {
		return false
	}

	return isTimeoutErrorRecursive(cause, depth+1, maxDepth)
}
