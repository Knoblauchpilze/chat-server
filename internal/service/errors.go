package service

import (
	"github.com/Knoblauchpilze/backend-toolkit/pkg/errors"
)

const (
	ErrInvalidName             errors.ErrorCode = 400
	ErrEmptyMessage            errors.ErrorCode = 401
	ErrUserNotInRoom           errors.ErrorCode = 402
	ErrLeavingRoomIsNotAllowed errors.ErrorCode = 403
)
