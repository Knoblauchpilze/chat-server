package clients

import (
	"context"
	"sync"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/logger"
	bterrors "github.com/Knoblauchpilze/chat-server/pkg/errors"
	"github.com/Knoblauchpilze/chat-server/pkg/messages"
	"github.com/coder/websocket"
	"github.com/google/uuid"
)

type Manager interface {
	OnConnect(conn *websocket.Conn) (bool, uuid.UUID)
	OnDisconnect(id uuid.UUID)
	OnReadError(id uuid.UUID, err error)

	messages.Dispatcher
}

type managerImpl struct {
	log       logger.Logger
	queue     messages.OutgoingQueue
	handshake Handshake

	lock    sync.RWMutex
	clients map[uuid.UUID]*websocket.Conn
}

type ManagerProps struct {
	Queue     messages.OutgoingQueue
	Handshake Handshake
	Log       logger.Logger
}

func NewManager(props ManagerProps) Manager {
	return &managerImpl{
		log:       props.Log,
		queue:     props.Queue,
		handshake: props.Handshake,
		clients:   make(map[uuid.UUID]*websocket.Conn),
	}
}

func (m *managerImpl) OnConnect(conn *websocket.Conn) (bool, uuid.UUID) {
	var connId uuid.UUID
	var handshakeErr error

	err := bterrors.SafeRunSync(func() {
		connId, handshakeErr = m.handshake.Perform(conn)
	})

	if err != nil {
		m.log.Warnf("OnConnect: panic while performing handshake: %v", err)
		return false, uuid.Nil
	} else if handshakeErr != nil {
		m.log.Warnf("OnConnect: handshake failed: %v", handshakeErr)
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
		m.log.Warnf("Broadcast: failed to broadcast message %s: %v", msg.Type(), err)
	}

	m.lock.RLock()
	defer m.lock.RUnlock()

	for id, conn := range m.clients {
		// TODO: We should probably have some synchronization mechanism here.
		// Or at least check if this is already handled.
		err := conn.Write(context.Background(), websocket.MessageBinary, encoded)
		if err != nil {
			m.log.Warnf("Broadcast: got error when sending %d byte(s) (%v) to %v: %v", len(encoded), msg.Type(), id, err)
		}
	}
}

func (m *managerImpl) BroadcastExcept(id uuid.UUID, msg messages.Message) {
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

		err := conn.Write(context.Background(), websocket.MessageBinary, encoded)
		if err != nil {
			m.log.Warnf("BroadcastExcept: got error when sending %d byte(s) (%v) to %v: %v", len(encoded), msg.Type(), clientId, err)
		}
	}
}

func (m *managerImpl) SendTo(id uuid.UUID, msg messages.Message) {
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

	err = conn.Write(context.Background(), websocket.MessageBinary, encoded)
	if err != nil {
		m.log.Warnf("SendTo: got error when sending %d byte(s) (%v) to %v: %v", len(encoded), msg.Type(), id, err)
	}
}
