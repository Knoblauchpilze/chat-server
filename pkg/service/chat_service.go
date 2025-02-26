package service

import (
	"github.com/Knoblauchpilze/backend-toolkit/pkg/logger"
	"github.com/Knoblauchpilze/chat-server/pkg/clients"
	"github.com/Knoblauchpilze/chat-server/pkg/messages"
)

type Chat interface {
	GenerateCallbacks() clients.Callbacks
	Start()
	Stop()
}

const incomingMessagesBufferSize = 5

type chatImpl struct {
	clientManager            clients.Manager
	messageParser            messages.Parser
	messageProcessingService MessageProcessingService
}

func NewChatService(log logger.Logger) Chat {
	queue := make(messages.Queue, incomingMessagesBufferSize)
	manager := clients.NewManager(queue, log)

	return &chatImpl{
		clientManager:            manager,
		messageParser:            messages.NewParser(queue),
		messageProcessingService: NewMessageProcessingService(queue, manager),
	}
}

func (c *chatImpl) GenerateCallbacks() clients.Callbacks {
	return clients.Callbacks{
		ConnectCallback:    c.clientManager.OnConnect,
		DisconnectCallback: c.clientManager.OnDisconnect,
		ReadErrorCallback:  c.clientManager.OnReadError,
		ReadDataCallback:   c.messageParser.OnReadData,
	}
}

func (c *chatImpl) Start() {
	c.messageProcessingService.Start()
}

func (c *chatImpl) Stop() {
	c.messageProcessingService.Stop()
}
