package tcp

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/logger"
	"github.com/Knoblauchpilze/chat-server/pkg/clients"
	"github.com/Knoblauchpilze/chat-server/pkg/connection"
	"github.com/Knoblauchpilze/chat-server/pkg/errors"
	"github.com/coder/websocket"
	"github.com/google/uuid"
)

type connectionManager interface {
	OnClientConnected(*websocket.Conn)
	Close()
}

type connectionData struct {
	lock     sync.Mutex
	listener connection.Listener
}

type managerImpl struct {
	log                   logger.Logger
	readTimeout           time.Duration
	incompleteDataTimeout time.Duration
	callbacks             clients.Callbacks

	accepting atomic.Bool
	lock      sync.Mutex
	listeners map[uuid.UUID]*connectionData
	wg        sync.WaitGroup
}

func newConnectionManager(config managerConfig, log logger.Logger) connectionManager {
	m := &managerImpl{
		log:                   log,
		readTimeout:           config.ReadTimeout,
		incompleteDataTimeout: config.IncompleteDataTimeout,
		callbacks:             config.Callbacks,
		listeners:             make(map[uuid.UUID]*connectionData),
	}

	m.accepting.Store(true)

	return m
}

func (m *managerImpl) OnClientConnected(conn *websocket.Conn) {
	var connId uuid.UUID
	var err error
	accepted := m.accepting.Load()

	if accepted {
		cb := func() {
			accepted, connId = m.callbacks.OnConnect(conn)
		}
		err = m.callCallbackAndLogError(cb, "Connect", connId)
	}

	if !accepted {
		m.log.Infof("OnConnect: denied incoming connection")
		// Voluntarily ignoring errors
		conn.Close(websocket.StatusNormalClosure, "connection denied")
	} else if err != nil {
		m.log.Infof("OnConnect: %v generated an error (err: %v)", connId, err)
		// Voluntarily ignoring errors
		conn.Close(websocket.StatusNormalClosure, "connection error")
	} else {
		m.log.Debugf("OnConnect: %v assigned to new connection", connId)
		m.registerListenerForConnection(connId, conn)
	}
}

func (m *managerImpl) Close() {
	if !m.accepting.CompareAndSwap(true, false) {
		return
	}

	// Copy all listeners to prevent dead locks in case one is
	// removed due to a disconnect or read error.
	allListeners := make(map[uuid.UUID]*connectionData)

	func() {
		defer m.lock.Unlock()
		m.lock.Lock()

		// https://stackoverflow.com/questions/23057785/how-to-deep-copy-a-map-and-then-clear-the-original
		for id, data := range m.listeners {
			allListeners[id] = data
		}

		clear(m.listeners)
	}()

	for _, data := range allListeners {
		data.Close()

		cb := func() {
			m.callbacks.OnDisconnect(data.listener.Id())
		}
		m.callCallbackAndLogError(cb, "Disconnect", data.listener.Id())

	}

	m.wg.Wait()
}

func (m *managerImpl) prepareListenerOptions(connId uuid.UUID) connection.ListenerOptions {
	return connection.ListenerOptions{
		Id:                    connId,
		ReadTimeout:           m.readTimeout,
		IncompleteDataTimeout: m.incompleteDataTimeout,
		Callbacks: connection.Callbacks{
			DisconnectCallback: func(id uuid.UUID) {
				m.onClientDisconnected(id)
			},
			ReadErrorCallback: func(id uuid.UUID, err error) {
				m.onReadError(id, err)
			},
			ReadDataCallback: func(id uuid.UUID, data []byte) int {
				return m.onReadData(id, data)
			},
		},
	}
}

func (m *managerImpl) registerListenerForConnection(
	connId uuid.UUID, conn *websocket.Conn,
) {
	opts := m.prepareListenerOptions(connId)
	listener := connection.New(conn, opts)

	func() {
		defer m.lock.Unlock()
		m.lock.Lock()

		m.listeners[listener.Id()] = &connectionData{
			listener: listener,
		}
	}()

	listener.Start()
}

func (m *managerImpl) onClientDisconnected(id uuid.UUID) {
	m.closeConnection(id, true)
}

func (m *managerImpl) onReadError(id uuid.UUID, err error) {
	m.log.Warnf("OnReadError: %v generated an error (err: %v)", id, err)

	cb := func() {
		m.callbacks.OnReadError(id, err)
	}
	m.callCallbackAndLogError(cb, "OnReadError", id)
	m.closeConnection(id, true)
}

func (m *managerImpl) onReadData(id uuid.UUID, data []byte) int {
	var processed int
	var keepAlive bool

	cb := func() {
		processed, keepAlive = m.callbacks.OnReadData(id, data)
	}
	err := m.callCallbackAndLogError(cb, "OnReadData", id)

	keepAlive = keepAlive && m.accepting.Load()

	if !keepAlive {
		m.closeConnection(id, true)
	} else if err != nil {
		m.log.Errorf("OnRead: %v generated an error (err: %v)", id, err)
		m.closeConnection(id, true)
	}

	return processed
}

func (m *managerImpl) closeConnection(id uuid.UUID, triggerDisconnect bool) {
	var ok bool
	var data *connectionData
	func() {
		defer m.lock.Unlock()
		m.lock.Lock()

		data, ok = m.listeners[id]
		delete(m.listeners, id)
	}()

	if ok {
		go data.Close()
		m.log.Debugf("Triggered removal of connection %v", data.listener.Id())
	}

	if triggerDisconnect && m.accepting.Load() {
		cb := func() {
			m.callbacks.OnDisconnect(id)
		}
		m.callCallbackAndLogError(cb, "Disconnect", id)
	}
}

func (m *managerImpl) callCallbackAndLogError(
	proc errors.Process,
	processName string,
	connId uuid.UUID,
) error {
	err := errors.SafeRunSync(proc)
	if err != nil {
		m.log.Warnf("%s callback failed for connection %v with err: %v", processName, connId, err)
	}
	return err
}

func (d *connectionData) Close() {
	defer d.lock.Unlock()
	d.lock.Lock()

	d.listener.Close()
}
