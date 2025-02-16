package tcp

import (
	"time"

	"github.com/Knoblauchpilze/chat-server/pkg/clients"
)

type AcceptorConfig struct {
	Port      uint16
	Callbacks ServerCallbacks
}

type ManagerConfig struct {
	ReadTimeout time.Duration
	Callbacks   clients.Callbacks
}

type ServerConfiguration struct {
	Port      uint16
	Callbacks clients.Callbacks
}
