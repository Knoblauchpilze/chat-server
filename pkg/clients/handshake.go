package clients

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/errors"
	"github.com/Knoblauchpilze/chat-server/pkg/repositories"
	"github.com/coder/websocket"
	"github.com/google/uuid"
)

type Handshake interface {
	Perform(*websocket.Conn) (uuid.UUID, error)
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

func (h *handshakeImpl) Perform(conn *websocket.Conn) (uuid.UUID, error) {
	id, err := h.tryWaitForUserId(conn)
	if err != nil {
		return id, err
	}

	fmt.Printf("received client id: \"%v\"\n", id)

	_, err = h.userRepo.Get(context.Background(), id)
	if err != nil {
		return id, wrapHandshakeFailureError(err)
	}

	return id, nil
}

func (h *handshakeImpl) tryWaitForUserId(conn *websocket.Conn) (uuid.UUID, error) {
	ctx, cancel := context.WithTimeout(context.Background(), h.timeout)
	defer cancel()

	msgType, data, err := conn.Read(ctx)

	msgType2, data2, err2 := conn.Read(ctx)

	fmt.Printf("received %v message, %d byte(s): \"%v\"\n", msgType, len(data), string(data))
	fmt.Printf("received %v message, %d byte(s): \"%v\", err: %v\n", msgType2, len(data2), string(data2), err2)

	var id uuid.UUID

	if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
		// It is okay to mark the connection as closed here: we allowed for some
		// time to send the client's id. If it fails then we consider that it's
		// the client's fault.
		// If it happens too often we could consider increasing the timeout or
		// implementing a retry mechanism.
		return uuid.Nil, newHandshakeTimeoutError()
	} else if err != nil {
		return uuid.Nil, wrapHandshakeFailureError(err)
	} else if len(data) != len(id) {
		return uuid.Nil, newHandshakeIncompleteError()
	}

	id, err = uuid.FromBytes(data)
	if err != nil {
		return uuid.Nil, errors.WrapCode(err, HandshakeFailure)
	}

	return id, nil
}
