package tcp

import (
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/logger"
	"github.com/Knoblauchpilze/chat-server/pkg/clients"
	"github.com/Knoblauchpilze/chat-server/pkg/connection"
	"github.com/Knoblauchpilze/chat-server/pkg/errors"
	"github.com/google/uuid"
)

type connectionManager interface {
	OnClientConnected(net.Conn)
	Close()
}

type connectionData struct {
	lock     sync.Mutex
	listener connection.Listener
}

type managerImpl struct {
	log         logger.Logger
	readTimeout time.Duration
	callbacks   clients.Callbacks

	accepting atomic.Bool
	lock      sync.Mutex
	listeners map[uuid.UUID]*connectionData
	wg        sync.WaitGroup
}

func newConnectionManager(config managerConfig, log logger.Logger) connectionManager {
	m := &managerImpl{
		log:         log,
		readTimeout: config.ReadTimeout,
		callbacks:   config.Callbacks,
		listeners:   make(map[uuid.UUID]*connectionData),
	}

	m.accepting.Store(true)

	return m
}

func (m *managerImpl) OnClientConnected(conn net.Conn) {
	opts := m.prepareListenerOptions()
	listener := connection.New(conn, opts)

	m.registerListener(listener)
	m.handleOnConnect(conn, listener)
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

func (m *managerImpl) prepareListenerOptions() connection.ListenerOptions {
	return connection.ListenerOptions{
		ReadTimeout: m.readTimeout,
		Callbacks: connection.Callbacks{
			DisconnectCallbacks: []connection.OnDisconnect{
				func(id uuid.UUID) {
					m.onClientDisconnected(id)
				},
			},
			ReadErrorCallbacks: []connection.OnReadError{
				func(id uuid.UUID, err error) {
					m.onReadError(id, err)
				},
			},
			ReadDataCallbacks: []connection.OnReadData{
				func(id uuid.UUID, data []byte) {
					m.onReadData(id, data)
				},
			},
		},
	}
}

func (m *managerImpl) registerListener(listener connection.Listener) {
	defer m.lock.Unlock()
	m.lock.Lock()

	m.listeners[listener.Id()] = &connectionData{
		listener: listener,
	}
}

func (m *managerImpl) handleOnConnect(conn net.Conn, listener connection.Listener) {
	var err error
	accepted := m.accepting.Load()
	connId := listener.Id()

	if accepted {
		cb := func() {
			accepted = m.callbacks.OnConnect(connId, conn)
		}
		err = m.callCallbackAndLogError(cb, "Connect", connId)
	}

	address := conn.RemoteAddr().String()
	if !accepted {
		m.log.Infof("OnConnect: denied connection from %v", address)
		m.closeConnection(connId, false)
	} else if err != nil {
		m.log.Infof("OnConnect: %v generated an error (err: %v)", connId, address, err)
		m.closeConnection(connId, false)
	} else {
		m.log.Debugf("OnConnect: %v assigned to %v", address, connId)
		listener.Start()
	}
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

func (m *managerImpl) onReadData(id uuid.UUID, data []byte) {
	var keepAlive bool

	cb := func() {
		keepAlive = m.callbacks.OnReadData(id, data)
	}
	err := m.callCallbackAndLogError(cb, "OnReadData", id)

	keepAlive = keepAlive && m.accepting.Load()

	if !keepAlive {
		m.closeConnection(id, true)
	} else if err != nil {
		m.log.Errorf("OnRead: %v generated an error (err: %v)", id, err)
		m.closeConnection(id, true)
	}
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
