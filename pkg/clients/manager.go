package clients

import (
	"sync"
	"sync/atomic"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/errors"
	"github.com/Knoblauchpilze/chat-server/pkg/messages"
	"github.com/Knoblauchpilze/chat-server/pkg/persistence"
	"github.com/google/uuid"
)

type Manager interface {
	Start() error
	Stop() error

	OnConnect(id uuid.UUID, client Client) error
	OnDisconnect(id uuid.UUID)

	messages.Dispatcher
}

type managerImpl struct {
	running atomic.Bool
	quit    chan struct{}
	done    chan struct{}

	lock    sync.RWMutex
	clients map[uuid.UUID]Client
}

func NewManager() Manager {
	return &managerImpl{
		quit: make(chan struct{}, 1),
		done: make(chan struct{}, 1),

		clients: make(map[uuid.UUID]Client),
	}
}

func (m *managerImpl) Start() error {
	if !m.running.CompareAndSwap(false, true) {
		return nil
	}

	<-m.quit
	defer func() {
		m.done <- struct{}{}
	}()

	var err error

	func() {
		m.lock.Lock()
		defer m.lock.Unlock()

		for _, client := range m.clients {
			clientErr := client.Stop()
			if clientErr != nil && err == nil {
				err = clientErr
			}
		}

		clear(m.clients)
	}()

	return err
}

func (m *managerImpl) Stop() error {
	if !m.running.CompareAndSwap(true, false) {
		return nil
	}

	m.quit <- struct{}{}
	<-m.done

	return nil
}

func (m *managerImpl) OnConnect(id uuid.UUID, client Client) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	if _, ok := m.clients[id]; ok {
		return errors.NewCode(ErrClientAlreadyRegistered)
	}

	m.clients[id] = client

	return nil
}

func (m *managerImpl) OnDisconnect(id uuid.UUID) {
	m.lock.Lock()
	defer m.lock.Unlock()

	delete(m.clients, id)
}

func (m *managerImpl) Broadcast(msg persistence.Message) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	for _, client := range m.clients {
		client.Enqueue(msg)
	}
}

func (m *managerImpl) BroadcastExcept(id uuid.UUID, msg persistence.Message) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	for clientId, client := range m.clients {
		if id == clientId {
			continue
		}

		client.Enqueue(msg)
	}
}

func (m *managerImpl) SendTo(id uuid.UUID, msg persistence.Message) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	client, ok := m.clients[id]
	if !ok {
		return
	}

	client.Enqueue(msg)
}
