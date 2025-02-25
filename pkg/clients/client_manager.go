package clients

import (
	"net"
	"sync"

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
	queue messages.Queue

	lock    sync.RWMutex
	clients map[uuid.UUID]net.Conn
}

func NewManager(queue messages.Queue) Manager {
	return &managerImpl{
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

	msg := messages.NewMessage(messages.CLIENT_CONNECTED)
	m.queue <- msg

	return true
}

func (m *managerImpl) OnDisconnect(id uuid.UUID) {
	func() {
		m.lock.Lock()
		defer m.lock.Unlock()

		delete(m.clients, id)
	}()

	msg := messages.NewMessage(messages.CLIENT_DISCONNECTED)
	m.queue <- msg
}

func (m *managerImpl) OnReadError(id uuid.UUID, err error) {
	func() {
		m.lock.Lock()
		defer m.lock.Unlock()

		delete(m.clients, id)
	}()

	msg := messages.NewMessage(messages.CLIENT_DISCONNECTED)
	m.queue <- msg
}

func (m *managerImpl) Broadcast(msg messages.Message) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	for _, conn := range m.clients {
		conn.Write([]byte("should broadcast something\n"))
	}
}

func (m *managerImpl) SendTo(id uuid.UUID, msg messages.Message) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	conn, ok := m.clients[id]
	if !ok {
		return
	}

	conn.Write([]byte("should write something\n"))
}
