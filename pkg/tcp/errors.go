package tcp

import "github.com/Knoblauchpilze/backend-toolkit/pkg/errors"

const (
	ErrTcpInitialization  errors.ErrorCode = 100
	ErrClientDisconnected errors.ErrorCode = 101
	ErrReadTimeout        errors.ErrorCode = 102
)
