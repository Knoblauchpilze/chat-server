package tcp

import (
	"github.com/coder/websocket"
)

type OnConnect func(conn *websocket.Conn)

type ServerCallbacks struct {
	ConnectCallback OnConnect
}

func (c ServerCallbacks) OnConnect(conn *websocket.Conn) {
	if c.ConnectCallback == nil {
		return
	}

	c.ConnectCallback(conn)
}
