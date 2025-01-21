package tcp

import "net"

type OnConnect func(net.Conn)

type ServerCallbacks struct {
	ConnectCallback OnConnect
	Connection      ConnectionCallbacks
}

func (c ServerCallbacks) OnConnect(conn net.Conn) {
	if c.ConnectCallback == nil {
		return
	}

	c.ConnectCallback(conn)
}
