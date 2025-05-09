package clients

import "github.com/Knoblauchpilze/backend-toolkit/pkg/errors"

const (
	ErrPartialSseWrite         errors.ErrorCode = 500
	ErrSseStreamFailed         errors.ErrorCode = 501
	ErrUnsupportedConnection   errors.ErrorCode = 502
	ErrClientAlreadyRegistered errors.ErrorCode = 503
	ErrBroadcastFailure        errors.ErrorCode = 504
)
