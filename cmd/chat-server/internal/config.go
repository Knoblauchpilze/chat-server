package internal

import (
	"github.com/Knoblauchpilze/chat-server/pkg/tcp"
)

type Configuration struct {
	Server tcp.Config
}

func DefaultConfig() Configuration {
	return Configuration{
		Server: tcp.Config{
			Port: uint16(80),
		},
	}
}
