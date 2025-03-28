package clients

import (
	"net"
	"time"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/errors"
	"github.com/google/uuid"
)

type HandshakeFunc func(net.Conn, time.Duration) (uuid.UUID, error)

func Handshake(conn net.Conn, timeout time.Duration) (uuid.UUID, error) {
	limit := time.Now().Add(timeout)
	conn.SetReadDeadline(limit)

	var id uuid.UUID
	data := make([]byte, len(id))

	n, err := conn.Read(data)
	if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
		return uuid.Nil, newHandshakeTimeoutError()
	} else if err != nil {
		return uuid.Nil, wrapHandshakeFailureError(err)
	} else if n != len(id) {
		return uuid.Nil, newHandshakeIncompleteError()
	}

	id, err = uuid.FromBytes(data)
	if err != nil {
		return uuid.Nil, errors.WrapCode(err, HandshakeFailure)
	}

	return id, nil
}
