package service

import (
	"time"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/logger"
	"github.com/Knoblauchpilze/chat-server/pkg/clients"
	"github.com/Knoblauchpilze/chat-server/pkg/messages"
)

type ChatService interface {
	GenerateCallbacks() clients.Callbacks
	Start()
	Stop()
}

const incomingMessagesBufferSize = 5

type chatServiceImpl struct {
	clientManager            clients.Manager
	messageParser            messages.Parser
	messageProcessingService messages.ProcessingService
}

func NewChatService(
	handshakeFunc clients.HandshakeFunc, connectTimeout time.Duration, log logger.Logger,
) ChatService {
	queue := make(chan messages.Message, incomingMessagesBufferSize)
	props := clients.ManagerProps{
		Queue:          queue,
		ConnectTimeout: connectTimeout,
		Handshake:      handshakeFunc,
		Log:            log,
	}
	manager := clients.NewManager(props)

	return &chatServiceImpl{
		clientManager:            manager,
		messageParser:            messages.NewParser(queue, log),
		messageProcessingService: messages.NewProcessingService(queue, manager, log),
	}
}

func (c *chatServiceImpl) GenerateCallbacks() clients.Callbacks {
	return clients.Callbacks{
		ConnectCallback:    c.clientManager.OnConnect,
		DisconnectCallback: c.clientManager.OnDisconnect,
		ReadErrorCallback:  c.clientManager.OnReadError,
		ReadDataCallback:   c.messageParser.OnReadData,
	}
}

func (c *chatServiceImpl) Start() {
	c.messageProcessingService.Start()
}

func (c *chatServiceImpl) Stop() {
	c.messageProcessingService.Stop()
}
