package clients

import "github.com/Knoblauchpilze/backend-toolkit/pkg/errors"

const (
	HandshakeTimeout    errors.ErrorCode = 500
	IncompleteHandshake errors.ErrorCode = 501
	HandshakeFailure    errors.ErrorCode = 502
)
