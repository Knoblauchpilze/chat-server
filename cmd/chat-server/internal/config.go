package internal

import (
	"time"

	"github.com/Knoblauchpilze/chat-server/pkg/clients"
)

type Configuration struct {
	Port        uint16
	ReadTimeout time.Duration
	Callbacks   clients.Callbacks
}

func DefaultConfig() Configuration {
	return Configuration{
		Port:        uint16(80),
		ReadTimeout: 3 * time.Second,
	}
}
