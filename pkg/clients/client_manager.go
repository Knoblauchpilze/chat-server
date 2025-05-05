package clients

import (
	"net"
	"sync"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/logger"
	"github.com/Knoblauchpilze/backend-toolkit/pkg/process"
	"github.com/Knoblauchpilze/chat-server/pkg/messages"
	"github.com/google/uuid"
)

type ClientManager interface {
	OnConnect(conn net.Conn) (bool, uuid.UUID)
	OnDisconnect(id uuid.UUID)
	OnReadError(id uuid.UUID, err error)

	messages.MessageDispatcher
}

type clientManagerImpl struct {
	log       logger.Logger
	queue     messages.OutgoingQueue
	handshake Handshake

	lock    sync.RWMutex
	clients map[uuid.UUID]net.Conn
}

type ManagerProps struct {
	Queue     messages.OutgoingQueue
	Handshake Handshake
	Log       logger.Logger
}

func NewClientManager(props ManagerProps) ClientManager {
	return &clientManagerImpl{
		log:       props.Log,
		queue:     props.Queue,
		handshake: props.Handshake,
		clients:   make(map[uuid.UUID]net.Conn),
	}
}

func (m *clientManagerImpl) OnConnect(conn net.Conn) (bool, uuid.UUID) {
	var connId uuid.UUID

	err := process.SafeRunSync(func() error {
		var err error
		connId, err = m.handshake.Perform(conn)
		return err
	})

	if err != nil {
		m.log.Warnf("OnConnect: handshake failed: %v", err)
		return false, uuid.Nil
	}

	func() {
		m.lock.Lock()
		defer m.lock.Unlock()

		m.clients[connId] = conn
	}()

	msg := messages.NewClientConnectedMessage(connId)
	m.queue <- msg

	return true, connId
}

func (m *clientManagerImpl) OnDisconnect(id uuid.UUID) {
	func() {
		m.lock.Lock()
		defer m.lock.Unlock()

		delete(m.clients, id)
	}()

	msg := messages.NewClientDisconnectedMessage(id)
	m.queue <- msg
}

func (m *clientManagerImpl) OnReadError(id uuid.UUID, err error) {
	func() {
		m.lock.Lock()
		defer m.lock.Unlock()

		delete(m.clients, id)
	}()

	msg := messages.NewClientDisconnectedMessage(id)
	m.queue <- msg
}

func (m *clientManagerImpl) Broadcast(msg messages.Message) {
	encoded, err := messages.Encode(msg)
	if err != nil {
		m.log.Warnf("Broadcast: failed to broadcast message %s: %v", msg.Type(), err)
	}

	m.lock.RLock()
	defer m.lock.RUnlock()

	for id, conn := range m.clients {
		// TODO: We should probably have some synchronization mechanism here.
		// Or at least check if this is already handled.
		n, err := conn.Write(encoded)
		if n != len(encoded) {
			m.log.Warnf("Broadcast: only sent %d byte(s) out of %d to %v", n, len(encoded), id)
		}
		if err != nil {
			m.log.Warnf("Broadcast: got error when sending %d byte(s) (%v) to %v: %v", len(encoded), msg.Type(), id, err)
		}
	}
}

func (m *clientManagerImpl) BroadcastExcept(id uuid.UUID, msg messages.Message) {
	encoded, err := messages.Encode(msg)
	if err != nil {
		m.log.Warnf("BroadcastExcept: failed to broadcast message %s: %v", msg.Type(), err)
	}

	m.lock.RLock()
	defer m.lock.RUnlock()

	for clientId, conn := range m.clients {
		if clientId == id {
			continue
		}

		n, err := conn.Write(encoded)
		if n != len(encoded) {
			m.log.Warnf("BroadcastExcept: only sent %d byte(s) out of %d to %v", n, len(encoded), clientId)
		}
		if err != nil {
			m.log.Warnf("BroadcastExcept: got error when sending %d byte(s) (%v) to %v: %v", len(encoded), msg.Type(), clientId, err)
		}
	}
}

func (m *clientManagerImpl) SendTo(id uuid.UUID, msg messages.Message) {
	encoded, err := messages.Encode(msg)
	if err != nil {
		m.log.Warnf("SendTo: failed to broadcast message %s to %v: %v", msg.Type(), id, err)
	}

	m.lock.RLock()
	defer m.lock.RUnlock()

	conn, ok := m.clients[id]
	if !ok {
		return
	}

	n, err := conn.Write(encoded)
	if n != len(encoded) {
		m.log.Warnf("SendTo: only sent %d byte(s) out of %d to %v", n, len(encoded), id)
	}
	if err != nil {
		m.log.Warnf("SendTo: got error when sending %d byte(s) (%v) to %v: %v", len(encoded), msg.Type(), id, err)
	}
}
