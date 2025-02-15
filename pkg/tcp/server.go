package tcp

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	bterr "github.com/Knoblauchpilze/backend-toolkit/pkg/errors"
	"github.com/Knoblauchpilze/backend-toolkit/pkg/logger"
	"github.com/Knoblauchpilze/chat-server/pkg/errors"
)

type Server interface {
	Start(ctx context.Context) error
}

// Inspiration for the shutdown mechanism:
// https://echo.labstack.com/docs/cookbook/graceful-shutdown
// https://eli.thegreenplace.net/2020/graceful-shutdown-of-a-tcp-server-in-go/

type serverImpl struct {
	port uint16

	log logger.Logger

	callbacks ServerCallbacks

	listener net.Listener
	running  atomic.Bool
	wg       sync.WaitGroup
}

func NewServer(config Config, log logger.Logger) Server {
	s := serverImpl{
		port:      config.Port,
		log:       log,
		callbacks: config.Callbacks,
	}

	return &s
}

func (s *serverImpl) Start(ctx context.Context) error {
	if !s.running.CompareAndSwap(false, true) {
		// Server is already running
		return bterr.NewCode(ErrAlreadyListening)
	}

	if err := s.initializeListener(); err != nil {
		return err
	}

	notifyCtx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	waitCtx, cancel := context.WithCancel(notifyCtx)

	var runError error

	s.wg.Add(1)

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

	fmt.Printf("shutting down\n")
	err := s.shutdown()
	fmt.Printf("waiting for shutting down finished\n")
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
	if !s.running.CompareAndSwap(true, false) {
		// Server is already shutting down.
		return nil
	}

	err := s.listener.Close()
	s.wg.Wait()
	return err
}

func (s *serverImpl) acceptLoop() error {
	defer s.wg.Done()

	running := true

	for running {
		fmt.Printf("accepting loop started\n")
		conn, err := s.listener.Accept()
		if err != nil {
			running = s.running.Load()
			if running {
				s.log.Errorf("Failed to accept connection in accept: %v", err)
			} else {
				fmt.Printf("accept loop finished\n")
			}
		}

		if running {
			fmt.Printf("received connection, accepting\n")
			s.wg.Add(1)
			go s.acceptConnection(conn)
		} else {
			fmt.Printf("can't accept server is not running anymore\n")
		}
	}

	fmt.Printf("exiting accept loop\n")

	return nil
}

func (s *serverImpl) acceptConnection(conn net.Conn) {
	defer s.wg.Done()

	fmt.Printf("calling onConnect callback\n")
	err := errors.SafeRunSync(
		func() {
			s.callbacks.OnConnect(conn)
		},
	)

	if err != nil {
		s.log.Warnf("Failed to accept connection from %v: %v", conn.RemoteAddr(), err)
		conn.Close()
	}

	fmt.Printf("callback called and finished\n")
}
