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
		// https://serverfault.com/questions/11806/which-ports-to-use-on-a-self-written-tcp-server
		Port: uint16(49152),
	}
}
