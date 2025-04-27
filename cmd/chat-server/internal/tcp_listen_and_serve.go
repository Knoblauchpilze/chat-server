package internal

import (
	"context"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/logger"
	"github.com/Knoblauchpilze/backend-toolkit/pkg/rest"
	"github.com/Knoblauchpilze/chat-server/pkg/tcp"
)

func tcpListenAndServe(
	ctx context.Context, config Configuration, log logger.Logger) error {
	endpoint := rest.ConcatenateEndpoints(config.Server.BasePath, config.TcpServer.BasePath)

	tcpConfig := tcp.ServerConfiguration{
		BasePath:        endpoint,
		Port:            config.TcpServer.Port,
		ShutdownTimeout: config.TcpServer.ShutdownTimeout,
		Callbacks:       config.Callbacks,
	}

	s, err := tcp.NewServer(tcpConfig, log)
	if err != nil {
		return err
	}

	return s.Start(ctx)
}
