package service

import (
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

func NewChatService(log logger.Logger) ChatService {
	queue := make(chan messages.Message, incomingMessagesBufferSize)
	manager := clients.NewManager(queue, log)

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
