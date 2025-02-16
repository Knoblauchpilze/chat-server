package tcp

import "github.com/Knoblauchpilze/backend-toolkit/pkg/errors"

const (
	ErrTcpInitialization errors.ErrorCode = 100
	ErrAlreadyListening  errors.ErrorCode = 101
	ErrAlreadyRunning    errors.ErrorCode = 102
)
