package clients

import "github.com/Knoblauchpilze/backend-toolkit/pkg/errors"

const (
	HandshakeTimeout    errors.ErrorCode = 500
	IncompleteHandshake errors.ErrorCode = 501
	HandshakeFailure    errors.ErrorCode = 502
)

func newHandshakeTimeoutError() error {
	return errors.NewCodeWithDetails(HandshakeTimeout, "timeout")
}

func newHandshakeIncompleteError() error {
	return errors.NewCodeWithDetails(
		IncompleteHandshake,
		"not enough data to received to perform handshake",
	)
}

func wrapHandshakeFailureError(err error) error {
	return errors.WrapCode(err, HandshakeFailure)
}
