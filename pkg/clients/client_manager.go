package clients

import (
	"net"

	"github.com/Knoblauchpilze/chat-server/pkg/messages"
	"github.com/google/uuid"
)

type Manager interface {
	OnConnect(id uuid.UUID, conn net.Conn) bool
	OnDisconnect(id uuid.UUID)
	OnReadError(id uuid.UUID, err error)
}

type managerImpl struct {
	queue messages.Queue
}

func NewManager(queue messages.Queue) Manager {
	return &managerImpl{
		queue: queue,
	}
}

func (m *managerImpl) OnConnect(id uuid.UUID, conn net.Conn) bool {
	msg := messages.NewMessage(messages.CLIENT_CONNECTED)
	m.queue <- msg
	return true
}

func (m *managerImpl) OnDisconnect(id uuid.UUID) {
	msg := messages.NewMessage(messages.CLIENT_DISCONNECTED)
	m.queue <- msg
}

func (m *managerImpl) OnReadError(id uuid.UUID, err error) {
	msg := messages.NewMessage(messages.CLIENT_DISCONNECTED)
	m.queue <- msg
}
