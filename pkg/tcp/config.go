package tcp

import (
	"time"

	"github.com/Knoblauchpilze/chat-server/pkg/clients"
)

type acceptorConfig struct {
	BasePath        string
	Port            uint16
	ShutdownTimeout time.Duration
	Callbacks       ServerCallbacks
}

type managerConfig struct {
	ReadTimeout           time.Duration
	IncompleteDataTimeout time.Duration
	Callbacks             clients.Callbacks
}

type ServerConfiguration struct {
	BasePath        string
	Port            uint16
	ShutdownTimeout time.Duration
	Callbacks       clients.Callbacks
}
