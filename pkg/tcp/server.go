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
	"github.com/Knoblauchpilze/backend-toolkit/pkg/server"
	"github.com/google/uuid"
)

type Server interface {
	Start(ctx context.Context) error
}

type serverImpl struct {
	port uint16

	log logger.Logger

	listener        net.Listener
	shutdownTimeout time.Duration
	quit            chan interface{}

	lock        sync.Mutex
	connections map[uuid.UUID]ConnectionListener
}

func NewServer(config server.Config, log logger.Logger) Server {
	return &serverImpl{
		port: config.Port,

		log: log,

		shutdownTimeout: config.ShutdownTimeout,
		quit:            make(chan interface{}),

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
	defer s.lock.Unlock()
	s.lock.Lock()

	for _, conn := range s.connections {
		conn.Close()
	}

	clear(s.connections)
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
	s.lock.Lock()
	defer s.lock.Unlock()

	opts := ConnectionListenerOptions{
		ReadTimeout: s.shutdownTimeout - 1*time.Second,
		Callbacks: ConnectionCallbacks{
			DisconnectCallbacks: []OnDisconnect{
				func(id uuid.UUID) {
					s.handleConnectionRemoval(id)
				},
			},
			PanicCallbacks: []OnPanic{
				func(id uuid.UUID, err error) {
					s.handleConnectionRemoval(id)
				},
			},
		},
	}
	listener := newListener(conn, opts)
	s.connections[listener.Id()] = listener

	s.log.Debugf("Registered connection %v", listener.Id())
	listener.StartListening()
}

func (s *serverImpl) handleConnectionRemoval(id uuid.UUID) {
	delete(s.connections, id)
	s.log.Debugf("Removed connection %v", id)
}
