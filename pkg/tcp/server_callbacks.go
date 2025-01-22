package tcp

import (
	"net"

	"github.com/google/uuid"
)

type OnConnect func(id uuid.UUID, conn net.Conn)

type ServerCallbacks struct {
	ConnectCallback OnConnect
	Connection      ConnectionCallbacks
}

func (c ServerCallbacks) OnConnect(id uuid.UUID, conn net.Conn) {
	if c.ConnectCallback == nil {
		return
	}

	c.ConnectCallback(id, conn)
}
