package internal

import (
	"context"
	"fmt"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/logger"
	"github.com/Knoblauchpilze/chat-server/pkg/tcp"
)

func tcpListenAndServe(
	ctx context.Context, config Configuration, log logger.Logger) error {
	tcpConfig := tcp.ServerConfiguration{
		// TODO: Make this configurable
		BasePath:        fmt.Sprintf("%s/ws", config.Server.BasePath),
		Port:            config.TcpPort,
		ShutdownTimeout: config.Server.ShutdownTimeout,
		Callbacks:       config.Callbacks,
	}

	s, err := tcp.NewServer(tcpConfig, log)
	if err != nil {
		return err
	}

	return s.Start(ctx)
}
