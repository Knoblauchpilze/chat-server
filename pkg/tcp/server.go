package tcp

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	bterr "github.com/Knoblauchpilze/backend-toolkit/pkg/errors"
	"github.com/Knoblauchpilze/backend-toolkit/pkg/logger"
	"github.com/google/uuid"
)

type Server interface {
	Start(ctx context.Context) error
}

type serverImpl struct {
	port uint16

	log logger.Logger

	listener                  net.Listener
	connectionShutdownTimeout time.Duration
	quit                      chan interface{}

	callbacks   ServerCallbacks
	lock        sync.Mutex
	connections map[uuid.UUID]ConnectionListener
}

func NewServer(config Config, log logger.Logger) Server {
	return &serverImpl{
		port: config.Port,

		log: log,

		connectionShutdownTimeout: config.ShutdownTimeout,
		quit:                      make(chan interface{}),

		callbacks:   config.Callbacks,
		connections: make(map[uuid.UUID]ConnectionListener),
	}
}

func (s *serverImpl) Start(ctx context.Context) error {
	if err := s.initializeListener(); err != nil {
		return err
	}

	// https://echo.labstack.com/docs/cookbook/graceful-shutdown
	notifyCtx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	waitCtx, cancel := context.WithCancel(notifyCtx)

	var runError error

	go func() {
		s.log.Infof("Starting server")
		err := s.acceptLoop()
		s.log.Infof("Server is shutting down")

		if err != nil {
			runError = err
			cancel()
		}
	}()

	const reasonableWaitTimeToInitializeServer = 50 * time.Millisecond
	time.Sleep(reasonableWaitTimeToInitializeServer)

	<-waitCtx.Done()

	err := s.shutdown()
	if err != nil {
		return err
	} else if runError != nil {
		return runError
	}

	s.log.Infof("Server gracefully shutdown")
	return nil
}

func (s *serverImpl) initializeListener() error {
	var err error

	address := fmt.Sprintf(":%d", s.port)
	s.listener, err = net.Listen("tcp", address)

	if err != nil {
		return bterr.WrapCode(err, ErrTcpInitialization)
	}

	s.log.Infof("Server will be listening at %s", address)

	return nil
}

func (s *serverImpl) shutdown() error {
	// https://eli.thegreenplace.net/2020/graceful-shutdown-of-a-tcp-server-in-go/
	close(s.quit)
	err := s.listener.Close()
	s.closeAllConnections()
	return err
}

func (s *serverImpl) closeAllConnections() {
	// Copy all connections to prevent dead locks in case one is
	// removed due to a disconnect or read error.
	// var allConnections map[uuid.UUID]ConnectionListener
	allConnections := make(map[uuid.UUID]ConnectionListener)

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

func (s *serverImpl) acceptLoop() error {
	running := true

	for running {
		accept := true

		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-s.quit:
				running = false
				accept = false
			default:
				s.log.Errorf("Failed to accept connection: %v", err)
				accept = false
			}
		}

		if accept {
			s.handleConnection(conn)
		}
	}

	return nil
}

func (s *serverImpl) handleConnection(conn net.Conn) {
	var listener ConnectionListener

	func() {
		defer s.lock.Unlock()
		s.lock.Lock()

		opts := ConnectionListenerOptions{
			ReadTimeout: s.connectionShutdownTimeout,
			Callbacks:   s.callbacks.Connection,
		}
		opts.Callbacks.DisconnectCallbacks = append(opts.Callbacks.DisconnectCallbacks,
			func(id uuid.UUID) {
				s.handleConnectionRemoval(id)
			})
		opts.Callbacks.PanicCallbacks = append(opts.Callbacks.PanicCallbacks,
			func(id uuid.UUID, err error) {
				s.handleConnectionRemoval(id)
			})

		listener = newListener(conn, opts)
		s.connections[listener.Id()] = listener
	}()

	s.callbacks.OnConnect(listener.Id(), conn)

	s.log.Debugf("Registered connection %v", listener.Id())
	listener.StartListening()
}

func (s *serverImpl) handleConnectionRemoval(id uuid.UUID) {
	s.lock.Lock()
	defer s.lock.Unlock()

	delete(s.connections, id)
	s.log.Debugf("Removed connection %v", id)
}
