package service

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Knoblauchpilze/chat-server/pkg/clients"
	"github.com/Knoblauchpilze/chat-server/pkg/messages"
)

type MessageProcessingService interface {
	Start()
	Stop()
}

const timeoutForMessageProcessing = 1 * time.Second

type messageProcessingServiceImpl struct {
	queue   messages.Queue
	manager clients.Manager

	running atomic.Bool
	wg      sync.WaitGroup
}

func NewMessageProcessingService(queue messages.Queue, manager clients.Manager) MessageProcessingService {
	return &messageProcessingServiceImpl{
		queue:   queue,
		manager: manager,
	}
}

func (m *messageProcessingServiceImpl) Start() {
	if !m.running.CompareAndSwap(false, true) {
		return
	}

	m.wg.Add(1)

	go m.activeLoop()
}

func (m *messageProcessingServiceImpl) Stop() {
	if !m.running.CompareAndSwap(true, false) {
		return
	}

	m.wg.Wait()
}

func (m *messageProcessingServiceImpl) activeLoop() {
	defer m.wg.Done()

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
	fmt.Printf("received message %v\n", msg.Type())
	m.manager.Broadcast(msg)
}
