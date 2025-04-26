package clients

import "github.com/Knoblauchpilze/backend-toolkit/pkg/errors"

const (
	ErrHandshakeTimeout    errors.ErrorCode = 500
	ErrIncompleteHandshake errors.ErrorCode = 501
	ErrHandshakeFailure    errors.ErrorCode = 502
)

func newHandshakeTimeoutError() error {
	return errors.NewCodeWithDetails(ErrHandshakeTimeout, "timeout")
}

func newHandshakeIncompleteError() error {
	return errors.NewCodeWithDetails(
		ErrIncompleteHandshake,
		"not enough data to received to perform handshake",
	)
}

func wrapHandshakeFailureError(err error) error {
	return errors.WrapCode(err, ErrHandshakeFailure)
}
