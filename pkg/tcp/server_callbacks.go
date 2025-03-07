package tcp

import "net"

type OnConnect func(conn net.Conn)

type ServerCallbacks struct {
	ConnectCallback OnConnect
}

func (c ServerCallbacks) OnConnect(conn net.Conn) {
	if c.ConnectCallback == nil {
		return
	}

	c.ConnectCallback(conn)
}
