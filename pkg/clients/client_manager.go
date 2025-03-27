package clients

import (
	"net"
	"sync"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/logger"
	"github.com/Knoblauchpilze/chat-server/pkg/messages"
	"github.com/google/uuid"
)

type Manager interface {
	OnConnect(conn net.Conn) (bool, uuid.UUID)
	OnDisconnect(id uuid.UUID)
	OnReadError(id uuid.UUID, err error)

	messages.Dispatcher
}

type managerImpl struct {
	log   logger.Logger
	queue messages.OutgoingQueue

	lock    sync.RWMutex
	clients map[uuid.UUID]net.Conn
}

func NewManager(queue messages.OutgoingQueue, log logger.Logger) Manager {
	return &managerImpl{
		log:     log,
		queue:   queue,
		clients: make(map[uuid.UUID]net.Conn),
	}
}

func (m *managerImpl) OnConnect(conn net.Conn) (bool, uuid.UUID) {
	connId := uuid.New()

	func() {
		m.lock.Lock()
		defer m.lock.Unlock()

		m.clients[connId] = conn
	}()

	msg := messages.NewClientConnectedMessage(connId)
	m.queue <- msg

	// TODO: We should have a handshake here and potentially deny clients.
	return true, connId
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

	for id, conn := range m.clients {
		// TODO: We should probably have some synchronization mechanism here.
		// Or at least check if this is already handled.
		n, err := conn.Write(encoded)
		if n != len(encoded) {
			m.log.Warnf("Only sent %d byte(s) out of %d to %v", n, len(encoded), id)
		}
		if err != nil {
			m.log.Warnf("Got error when sending %d byte(s) to %v: %v", len(encoded), id, err)
		}
	}
}

func (m *managerImpl) BroadcastExcept(id uuid.UUID, msg messages.Message) {
	encoded, err := messages.Encode(msg)
	if err != nil {
		m.log.Warnf("Failed to broadcast message %s: %v", msg.Type(), err)
	}

	m.lock.RLock()
	defer m.lock.RUnlock()

	for clientId, conn := range m.clients {
		if clientId == id {
			continue
		}

		n, err := conn.Write(encoded)
		if n != len(encoded) {
			m.log.Warnf("Only sent %d byte(s) out of %d to %v", n, len(encoded), clientId)
		}
		if err != nil {
			m.log.Warnf("Got error when sending %d byte(s) to %v: %v", len(encoded), clientId, err)
		}
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

	n, err := conn.Write(encoded)
	if n != len(encoded) {
		m.log.Warnf("Only sent %d byte(s) out of %d to %v", n, len(encoded), id)
	}
	if err != nil {
		m.log.Warnf("Got error when sending %d byte(s) to %v: %v", len(encoded), id, err)
	}
}
