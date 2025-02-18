package internal

import (
	"context"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/logger"
	"github.com/Knoblauchpilze/chat-server/pkg/clients"
	"github.com/Knoblauchpilze/chat-server/pkg/tcp"
	"github.com/google/uuid"
)

type Server interface {
	Start(context.Context) error
}

type serverImpl struct {
	log       logger.Logger
	server    tcp.Server
	callbacks clients.Callbacks
}

func NewServer(config Configuration, log logger.Logger) (Server, error) {
	s := &serverImpl{
		log:       log,
		callbacks: config.Callbacks,
	}

	tcpConfig := tcp.ServerConfiguration{
		Port: config.Port,
		Callbacks: clients.Callbacks{
			ConnectCallback: func(id uuid.UUID, address string) bool {
				return s.onConnect(id, address)
			},
			DisconnectCallback: func(id uuid.UUID) {
				s.onDisconnect(id)
			},
			ReadErrorCallback: func(id uuid.UUID, err error) {
				s.onReadError(id, err)
			},
			ReadDataCallback: func(id uuid.UUID, data []byte) bool {
				return s.onReadData(id, data)
			},
		},
	}

	var err error
	s.server, err = tcp.NewServer(tcpConfig, log)

	return s, err
}

func (s *serverImpl) Start(ctx context.Context) error {
	return s.server.Start(ctx)
}

func (s *serverImpl) onConnect(id uuid.UUID, address string) bool {
	return s.callbacks.OnConnect(id, address)
}

func (s *serverImpl) onDisconnect(id uuid.UUID) {
	s.callbacks.OnDisconnect(id)
}

func (s *serverImpl) onReadData(id uuid.UUID, data []byte) bool {
	return s.callbacks.OnReadData(id, data)
}

func (s *serverImpl) onReadError(id uuid.UUID, err error) {
	s.callbacks.OnReadError(id, err)
}
