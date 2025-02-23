package clients

import (
	"net"

	"github.com/google/uuid"
)

// The return value should indicate whether or not the connection is accepted.
type OnConnect func(id uuid.UUID, conn net.Conn) bool

// The connection will automatically be closed after this callback is triggered.
type OnDisconnect func(id uuid.UUID)

// The return value should indicate whether or not the connection should stay open.
type OnReadData func(id uuid.UUID, data []byte) bool

// The connection will automatically be closed after this callback is triggered.
type OnReadError func(id uuid.UUID, err error)

type Callbacks struct {
	ConnectCallback    OnConnect
	DisconnectCallback OnDisconnect
	ReadErrorCallback  OnReadError
	ReadDataCallback   OnReadData
}

func (c Callbacks) OnConnect(id uuid.UUID, conn net.Conn) bool {
	if c.ConnectCallback == nil {
		return true
	}
	return c.ConnectCallback(id, conn)
}

func (c Callbacks) OnDisconnect(id uuid.UUID) {
	if c.DisconnectCallback != nil {
		c.DisconnectCallback(id)
	}
}

func (c Callbacks) OnReadData(id uuid.UUID, data []byte) bool {
	if c.ReadDataCallback == nil {
		return true
	}
	return c.ReadDataCallback(id, data)
}

func (c Callbacks) OnReadError(id uuid.UUID, err error) {
	if c.ReadErrorCallback != nil {
		c.ReadErrorCallback(id, err)
	}
}
