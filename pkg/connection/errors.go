package connection

import "github.com/Knoblauchpilze/backend-toolkit/pkg/errors"

const (
	ErrClientDisconnected     errors.ErrorCode = 200
	ErrReadTimeout            errors.ErrorCode = 201
	ErrTooLargeIncompleteData errors.ErrorCode = 202
)
