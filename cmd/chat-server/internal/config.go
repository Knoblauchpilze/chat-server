package internal

import (
	"github.com/Knoblauchpilze/chat-server/pkg/clients"
)

type Configuration struct {
	Port      uint16
	Callbacks clients.Callbacks
}

func DefaultConfig() Configuration {
	return Configuration{
		Port: uint16(80),
	}
}
