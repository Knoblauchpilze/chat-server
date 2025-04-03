package tcp

import (
	"context"
	"net"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	bterr "github.com/Knoblauchpilze/backend-toolkit/pkg/errors"
	"github.com/Knoblauchpilze/backend-toolkit/pkg/logger"
)

const (
	connectionReadTimeout           = 1 * time.Second
	connectionIncompleteDataTimeout = 5 * time.Second
)

type Server interface {
	Start(ctx context.Context) error
}

type serverImpl struct {
	log      logger.Logger
	manager  connectionManager
	acceptor connectionAcceptor

	running atomic.Bool
	wg      sync.WaitGroup
}

func NewServer(config ServerConfiguration, log logger.Logger) (Server, error) {
	s := serverImpl{
		log: log,
	}

	managerConfig := managerConfig{
		ReadTimeout:           connectionReadTimeout,
		IncompleteDataTimeout: connectionIncompleteDataTimeout,
		Callbacks:             config.Callbacks,
	}
	s.manager = newConnectionManager(managerConfig, s.log)

	acceptorConfig := acceptorConfig{
		Port: config.Port,
		Callbacks: ServerCallbacks{
			ConnectCallback: func(conn net.Conn) {
				s.manager.OnClientConnected(conn)
			},
		},
	}
	var err error
	s.acceptor, err = newConnectionAcceptor(acceptorConfig, s.log)

	return &s, err
}

func (s *serverImpl) Start(ctx context.Context) error {
	if !s.running.CompareAndSwap(false, true) {
		// Server is already running.
		return bterr.NewCode(ErrAlreadyRunning)
	}

	notifyCtx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	waitCtx, cancel := context.WithCancel(notifyCtx)

	var runError error

	s.wg.Add(1)

	go func() {
		defer s.wg.Done()

		s.log.Infof("Starting server")
		err := s.acceptor.Accept()
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

func (s *serverImpl) shutdown() error {
	if !s.running.CompareAndSwap(true, false) {
		// Server is already shutting down.
		return nil
	}

	err := s.acceptor.Close()
	s.manager.Close()
	return err
}
