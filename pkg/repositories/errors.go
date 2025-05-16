package repositories

import "github.com/Knoblauchpilze/backend-toolkit/pkg/errors"

const (
	ErrNoSuchRoom                  errors.ErrorCode = 600
	ErrUserNotRegisteredInRoom     errors.ErrorCode = 601
	ErrNoSuchUser                  errors.ErrorCode = 602
	ErrUserAlreadyRegisteredInRoom errors.ErrorCode = 603
)
