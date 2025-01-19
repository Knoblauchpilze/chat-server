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

	"github.com/KnoblauchPilze/backend-toolkit/pkg/errors"
	"github.com/KnoblauchPilze/backend-toolkit/pkg/logger"
	"github.com/KnoblauchPilze/backend-toolkit/pkg/server"
)

type Server interface {
	Start(ctx context.Context) error
}

type serverImpl struct {
	basePath string
	port     uint16
	address  string

	log logger.Logger

	listener        net.Listener
	shutdownTimeout time.Duration
	quit            chan interface{}

	lock    sync.Mutex
	closers []ConnectionCloser
}

func NewServer(config server.Config, log logger.Logger) Server {
	return &serverImpl{
		basePath: config.BasePath,
		port:     config.Port,

		log: log,

		shutdownTimeout: config.ShutdownTimeout,
		quit:            make(chan interface{}),
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
		s.log.Infof("Starting server at %s", s.address)
		err := s.acceptLoop()
		s.log.Infof("Server at %s is shutting down", s.address)

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

	s.log.Infof("Server at %s gracefully shutdown", s.address)
	return nil
}

func (s *serverImpl) initializeListener() error {
	var err error

	s.address = fmt.Sprintf(":%d", s.port)
	s.listener, err = net.Listen("tcp", s.address)

	if err != nil {
		return errors.WrapCode(err, ErrTcpInitialization)
	}
	return nil
}

func (s *serverImpl) shutdown() error {
	// https://eli.thegreenplace.net/2020/graceful-shutdown-of-a-tcp-server-in-go/
	close(s.quit)
	err := s.listener.Close()

	func() {
		defer s.lock.Unlock()
		s.lock.Lock()

		for _, closer := range s.closers {
			closer()
		}
	}()

	return err
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

	opts := ConnectionHandlerOptions{
		ReadTimeout: s.shutdownTimeout - 1*time.Second,
	}
	closer, _ := HandleConnection(conn, opts)

	// TODO: We never clean the handler.
	s.closers = append(s.closers, closer)
}
