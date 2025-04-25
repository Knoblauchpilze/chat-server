package clients

import (
	"github.com/coder/websocket"
	"github.com/google/uuid"
)

// The return value should indicate whether or not the connection is accepted
// and an identifier to refer to it.
type OnConnect func(conn *websocket.Conn) (bool, uuid.UUID)

// The connection will automatically be closed after this callback is triggered.
type OnDisconnect func(id uuid.UUID)

// The return value should indicate whether or not the connection should stay open
// and how many bytes were processed.
type OnReadData func(id uuid.UUID, data []byte) (int, bool)

// The connection will automatically be closed after this callback is triggered.
type OnReadError func(id uuid.UUID, err error)

type Callbacks struct {
	ConnectCallback    OnConnect
	DisconnectCallback OnDisconnect
	ReadErrorCallback  OnReadError
	ReadDataCallback   OnReadData
}

func (c Callbacks) OnConnect(conn *websocket.Conn) (bool, uuid.UUID) {
	if c.ConnectCallback == nil {
		return false, uuid.Nil
	}
	return c.ConnectCallback(conn)
}

func (c Callbacks) OnDisconnect(id uuid.UUID) {
	if c.DisconnectCallback != nil {
		c.DisconnectCallback(id)
	}
}

func (c Callbacks) OnReadData(id uuid.UUID, data []byte) (int, bool) {
	if c.ReadDataCallback == nil {
		return 0, true
	}
	return c.ReadDataCallback(id, data)
}

func (c Callbacks) OnReadError(id uuid.UUID, err error) {
	if c.ReadErrorCallback != nil {
		c.ReadErrorCallback(id, err)
	}
}
