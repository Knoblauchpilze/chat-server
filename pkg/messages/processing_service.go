package messages

import (
	"sync/atomic"
	"time"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/errors"
	"github.com/Knoblauchpilze/backend-toolkit/pkg/logger"
)

type ProcessingService interface {
	Start()
	Stop()
}

const timeoutForMessageProcessing = 1 * time.Second

type processingServiceImpl struct {
	queue      IncomingQueue
	dispatcher MessageDispatcher

	log logger.Logger

	running atomic.Bool
	quit    chan struct{}
}

func NewProcessingService(
	queue IncomingQueue,
	dispatcher MessageDispatcher,
	log logger.Logger,
) ProcessingService {
	return &processingServiceImpl{
		queue:      queue,
		dispatcher: dispatcher,
		log:        log,
		quit:       make(chan struct{}, 1),
	}
}

func (m *processingServiceImpl) Start() {
	if !m.running.CompareAndSwap(false, true) {
		return
	}

	go m.activeLoop()
}

func (m *processingServiceImpl) Stop() {
	if !m.running.CompareAndSwap(true, false) {
		return
	}

	<-m.quit
}

func (m *processingServiceImpl) activeLoop() {
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

func (m *processingServiceImpl) waitForMessageOrTimeout() (Message, bool) {
	timeoutChan := time.After(timeoutForMessageProcessing)

	timeout := false
	var msg Message

	select {
	case <-timeoutChan:
		timeout = true
	case msg = <-m.queue:
	}

	return msg, timeout
}

func (m *processingServiceImpl) processMessage(msg Message) {
	var err error

	switch msg.Type() {
	case CLIENT_CONNECTED:
		err = m.processClientConnectedMessage(msg)
	case CLIENT_DISCONNECTED:
		m.dispatcher.Broadcast(msg)
	case DIRECT_MESSAGE:
		err = m.processDirectMessage(msg)
	case ROOM_MESSAGE:
		// TODO: Handle room message, we need a list of participants first
		err = errors.NotImplemented()
	default:
		err = errors.NewCode(UnrecognizedMessageType)
	}

	if err != nil {
		m.log.Warnf("Failed to process message with type %v, err: %v", msg.Type(), err)
	}
}

func (m *processingServiceImpl) processClientConnectedMessage(in Message) error {
	msg, err := ToMessageStruct[ClientConnectedMessage](in)
	if err != nil {
		return err
	}

	m.dispatcher.BroadcastExcept(msg.Client, msg)
	return nil
}

func (m *processingServiceImpl) processDirectMessage(in Message) error {
	msg, err := ToMessageStruct[DirectMessage](in)
	if err != nil {
		return err
	}

	m.dispatcher.SendTo(msg.Receiver, msg)
	return nil
}
