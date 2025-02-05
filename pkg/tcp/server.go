package tcp

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	bterr "github.com/Knoblauchpilze/backend-toolkit/pkg/errors"
	"github.com/Knoblauchpilze/backend-toolkit/pkg/logger"
)

type Server interface {
	Start(ctx context.Context) error
}

type serverImpl struct {
	port uint16

	log logger.Logger

	listener  net.Listener
	accepting atomic.Bool
	quit      chan interface{}

	callbacks ServerCallbacks
}

func NewServer(config Config, log logger.Logger) Server {
	s := serverImpl{
		port: config.Port,

		log: log,

		quit: make(chan interface{}),

		callbacks: config.Callbacks,
	}

	s.accepting.Store(true)

	return &s
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
	if !s.accepting.CompareAndSwap(true, false) {
		// Server already closed
		return nil
	}

	// https://eli.thegreenplace.net/2020/graceful-shutdown-of-a-tcp-server-in-go/
	close(s.quit)
	return s.listener.Close()
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
			s.callbacks.OnConnect(conn)
		}
	}

	return nil
}
