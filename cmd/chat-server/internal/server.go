package internal

import (
	"context"
	"net"
	"sync"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/logger"
	"github.com/Knoblauchpilze/chat-server/pkg/connection"
	"github.com/Knoblauchpilze/chat-server/pkg/tcp"
	"github.com/google/uuid"
)

type Server interface {
	Start(context.Context) error
}

type serverImpl struct {
	log    logger.Logger
	server tcp.Server

	lock        sync.Mutex
	connections map[uuid.UUID]connection.Listener
}

func NewServer(config Configuration, log logger.Logger) Server {
	s := &serverImpl{
		log:         log,
		connections: make(map[uuid.UUID]connection.Listener),
	}

	tcpConf := tcp.Config{
		Port: config.Port,
		Callbacks: tcp.ServerCallbacks{
			ConnectCallback: func(conn net.Conn) {
				s.onClientConnected(conn)
			},
		},
	}

	s.server = tcp.NewServer(tcpConf, log)
	return s
}

func (s *serverImpl) Start(ctx context.Context) error {
	err := s.server.Start(ctx)
	s.closeAllConnections()
	return err
}

func (s *serverImpl) closeAllConnections() {
	// Copy all connections to prevent dead locks in case one is
	// removed due to a disconnect or read error.
	allConnections := make(map[uuid.UUID]connection.Listener)

	func() {
		defer s.lock.Unlock()
		s.lock.Lock()

		// https://stackoverflow.com/questions/23057785/how-to-deep-copy-a-map-and-then-clear-the-original
		for id, conn := range s.connections {
			allConnections[id] = conn
		}

		clear(s.connections)
	}()

	for _, conn := range allConnections {
		conn.Close()
	}
}

func (s *serverImpl) onClientConnected(conn net.Conn) {
	opts := connection.ListenerOptions{
		Callbacks: connection.Callbacks{
			DisconnectCallbacks: []connection.OnDisconnect{
				func(id uuid.UUID) {
					s.onClientDisconnected(id)
				},
			},
			ReadErrorCallbacks: []connection.OnReadError{
				func(id uuid.UUID, err error) {
					s.onReadError(id, err)
				},
			},
			ReadDataCallbacks: []connection.OnReadData{
				func(id uuid.UUID, data []byte) {
					s.onReadData(id, data)
				},
			},
		},
	}

	listener := connection.New(conn, opts)

	func() {
		defer s.lock.Unlock()
		s.lock.Lock()

		s.connections[listener.Id()] = listener
	}()

	s.log.Debugf("Registered connection %v from %v ", listener.Id(), conn.RemoteAddr())
	listener.Start()
}

func (s *serverImpl) onClientDisconnected(id uuid.UUID) {
	s.lock.Lock()
	defer s.lock.Unlock()

	delete(s.connections, id)
	s.log.Debugf("Removed connection %v", id)
}

func (s *serverImpl) onReadError(id uuid.UUID, err error) {
	s.log.Warnf("Received read error from %v (err: %v)", id, err)
}

func (s *serverImpl) onReadData(id uuid.UUID, data []byte) {
	s.log.Infof("Received data from %v: %s", id, string(data))
}
