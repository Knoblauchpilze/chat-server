package clients

import (
	"net"
	"time"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/errors"
	"github.com/Knoblauchpilze/chat-server/pkg/repositories"
	"github.com/google/uuid"
)

type Handshake interface {
	Perform(net.Conn) (uuid.UUID, error)
}

type handshakeImpl struct {
	userRepo repositories.UserRepository
	timeout  time.Duration
}

func NewHandshake(
	userRepo repositories.UserRepository, timeout time.Duration,
) Handshake {
	return &handshakeImpl{
		userRepo: userRepo,
		timeout:  timeout,
	}
}

func (h *handshakeImpl) Perform(conn net.Conn) (uuid.UUID, error) {
	limit := time.Now().Add(h.timeout)
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
