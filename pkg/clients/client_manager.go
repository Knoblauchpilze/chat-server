package clients

import (
	"net"
	"sync"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/logger"
	"github.com/Knoblauchpilze/chat-server/pkg/messages"
	"github.com/google/uuid"
)

type Manager interface {
	OnConnect(id uuid.UUID, conn net.Conn) bool
	OnDisconnect(id uuid.UUID)
	OnReadError(id uuid.UUID, err error)

	Broadcast(msg messages.Message)
	SendTo(id uuid.UUID, msg messages.Message)
}

type managerImpl struct {
	log   logger.Logger
	queue messages.Queue

	lock    sync.RWMutex
	clients map[uuid.UUID]net.Conn
}

func NewManager(queue messages.Queue, log logger.Logger) Manager {
	return &managerImpl{
		log:     log,
		queue:   queue,
		clients: make(map[uuid.UUID]net.Conn),
	}
}

func (m *managerImpl) OnConnect(id uuid.UUID, conn net.Conn) bool {
	func() {
		m.lock.Lock()
		defer m.lock.Unlock()

		m.clients[id] = conn
	}()

	msg := messages.NewClientConnectedMessage(id)
	m.queue <- msg

	return true
}

func (m *managerImpl) OnDisconnect(id uuid.UUID) {
	func() {
		m.lock.Lock()
		defer m.lock.Unlock()

		delete(m.clients, id)
	}()

	msg := messages.NewClientDisconnectedMessage(id)
	m.queue <- msg
}

func (m *managerImpl) OnReadError(id uuid.UUID, err error) {
	func() {
		m.lock.Lock()
		defer m.lock.Unlock()

		delete(m.clients, id)
	}()

	msg := messages.NewClientDisconnectedMessage(id)
	m.queue <- msg
}

func (m *managerImpl) Broadcast(msg messages.Message) {
	encoded, err := messages.Encode(msg)
	if err != nil {
		m.log.Warnf("Failed to broadcast message %s: %v", msg.Type(), err)
	}

	m.lock.RLock()
	defer m.lock.RUnlock()

	for _, conn := range m.clients {
		conn.Write(encoded)
	}
}

func (m *managerImpl) SendTo(id uuid.UUID, msg messages.Message) {
	encoded, err := messages.Encode(msg)
	if err != nil {
		m.log.Warnf("Failed to broadcast message %s to %v: %v", msg.Type(), id, err)
	}

	m.lock.RLock()
	defer m.lock.RUnlock()

	conn, ok := m.clients[id]
	if !ok {
		return
	}

	conn.Write(encoded)
}
