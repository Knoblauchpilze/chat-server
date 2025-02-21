package internal

import (
	"context"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/logger"
	"github.com/Knoblauchpilze/chat-server/pkg/clients"
	"github.com/Knoblauchpilze/chat-server/pkg/tcp"
	"github.com/google/uuid"
)

func ListenAndServe(ctx context.Context,
	config Configuration, log logger.Logger) error {
	tcpConfig := tcp.ServerConfiguration{
		Port: config.Port,
		Callbacks: clients.Callbacks{
			ConnectCallback: func(id uuid.UUID, address string) bool {
				return config.Callbacks.OnConnect(id, address)
			},
			DisconnectCallback: func(id uuid.UUID) {
				config.Callbacks.OnDisconnect(id)
			},
			ReadErrorCallback: func(id uuid.UUID, err error) {
				config.Callbacks.OnReadError(id, err)
			},
			ReadDataCallback: func(id uuid.UUID, data []byte) bool {
				return config.Callbacks.OnReadData(id, data)
			},
		},
	}

	s, err := tcp.NewServer(tcpConfig, log)
	if err != nil {
		return err
	}

	return s.Start(ctx)
}
