package messages

import "github.com/Knoblauchpilze/backend-toolkit/pkg/errors"

const (
	ErrUnrecognizedMessageFormat         errors.ErrorCode = 300
	ErrUnsupportedMessageType            errors.ErrorCode = 301
	ErrMessageDecodingFailed             errors.ErrorCode = 302
	ErrUnrecognizedMessageImplementation errors.ErrorCode = 303
	ErrMessageEncodingFailed             errors.ErrorCode = 304
)
