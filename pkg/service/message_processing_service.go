package service

import (
	"sync/atomic"
	"time"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/errors"
	"github.com/Knoblauchpilze/backend-toolkit/pkg/logger"
	"github.com/Knoblauchpilze/chat-server/pkg/clients"
	"github.com/Knoblauchpilze/chat-server/pkg/messages"
)

type MessageProcessingService interface {
	Start()
	Stop()
}

const timeoutForMessageProcessing = 1 * time.Second

type messageProcessingServiceImpl struct {
	queue   messages.IncomingQueue
	manager clients.Manager

	log logger.Logger

	running atomic.Bool
	quit    chan struct{}
}

func NewMessageProcessingService(
	queue messages.IncomingQueue,
	manager clients.Manager,
	log logger.Logger,
) MessageProcessingService {
	return &messageProcessingServiceImpl{
		queue:   queue,
		manager: manager,
		log:     log,
		quit:    make(chan struct{}, 1),
	}
}

func (m *messageProcessingServiceImpl) Start() {
	if !m.running.CompareAndSwap(false, true) {
		return
	}

	go m.activeLoop()
}

func (m *messageProcessingServiceImpl) Stop() {
	if !m.running.CompareAndSwap(true, false) {
		return
	}

	<-m.quit
}

func (m *messageProcessingServiceImpl) activeLoop() {
	defer func() {
		m.quit <- struct{}{}
	}()

	running := true
	for running {
		msg, timeout := m.waitForMessageOrTimeout()

		running = m.running.Load()

		if !timeout {
			m.processMessage(msg)
		}
	}
}

func (m *messageProcessingServiceImpl) waitForMessageOrTimeout() (messages.Message, bool) {
	timeoutChan := time.After(timeoutForMessageProcessing)

	timeout := false
	var msg messages.Message

	select {
	case <-timeoutChan:
		timeout = true
	case msg = <-m.queue:
	}

	return msg, timeout
}

func (m *messageProcessingServiceImpl) processMessage(msg messages.Message) {
	var err error

	switch msg.Type() {
	case messages.CLIENT_CONNECTED:
		err = m.processClientConnectedMessage(msg)
	case messages.CLIENT_DISCONNECTED:
		m.manager.Broadcast(msg)
	case messages.DIRECT_MESSAGE:
		err = m.processDirectMessage(msg)
	case messages.ROOM_MESSAGE:
		// TODO: Handle room message, we need a list of participants first
		err = errors.NotImplemented()
	default:
		err = errors.NewCode(UnrecognizedMessageType)
	}

	if err != nil {
		m.log.Warnf("Failed to process message with type %v, err: %v", msg.Type(), err)
	}
}

func (m *messageProcessingServiceImpl) processClientConnectedMessage(in messages.Message) error {
	msg, err := messages.ToMessageStruct[messages.ClientConnectedMessage](in)
	if err != nil {
		return err
	}

	m.manager.BroadcastExcept(msg.Client, msg)
	return nil
}

func (m *messageProcessingServiceImpl) processDirectMessage(in messages.Message) error {
	msg, err := messages.ToMessageStruct[messages.DirectMessage](in)
	if err != nil {
		return err
	}

	m.manager.SendTo(msg.Receiver, msg)
	return nil
}
