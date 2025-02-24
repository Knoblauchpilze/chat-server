package internal

import (
	"context"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/logger"
	"github.com/Knoblauchpilze/chat-server/pkg/tcp"
)

func ListenAndServe(ctx context.Context,
	config Configuration, log logger.Logger) error {
	tcpConfig := tcp.ServerConfiguration{
		Port:      config.Port,
		Callbacks: config.Callbacks,
	}

	s, err := tcp.NewServer(tcpConfig, log)
	if err != nil {
		return err
	}

	return s.Start(ctx)
}
