package internal

import (
	"time"

	"github.com/KnoblauchPilze/backend-toolkit/pkg/server"
)

type Configuration struct {
	Server server.Config
}

func DefaultConfig() Configuration {
	return Configuration{
		Server: server.Config{
			BasePath:        "/v1/chat-server",
			Port:            uint16(80),
			ShutdownTimeout: 5 * time.Second,
		},
	}
}
