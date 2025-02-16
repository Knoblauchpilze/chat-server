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

type ConnectionAcceptor interface {
	Listen(ctx context.Context) error
}

// Inspiration for the shutdown mechanism:
// https://echo.labstack.com/docs/cookbook/graceful-shutdown
// https://eli.thegreenplace.net/2020/graceful-shutdown-of-a-tcp-server-in-go/

type acceptorImpl struct {
	port uint16

	log logger.Logger

	callbacks ServerCallbacks

	listener net.Listener
	running  atomic.Bool
	wg       sync.WaitGroup
}

func NewConnectionAcceptor(config AcceptorConfig, log logger.Logger) ConnectionAcceptor {
	a := acceptorImpl{
		port:      config.Port,
		log:       log,
		callbacks: config.Callbacks,
	}

	return &a
}

func (a *acceptorImpl) Listen(ctx context.Context) error {
	if !a.running.CompareAndSwap(false, true) {
		// Server is already running
		return bterr.NewCode(ErrAlreadyListening)
	}

	if err := a.initializeListener(); err != nil {
		return err
	}

	notifyCtx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	waitCtx, cancel := context.WithCancel(notifyCtx)

	var runError error

	a.wg.Add(1)

	go func() {
		a.log.Infof("Starting server")
		err := a.acceptLoop()
		a.log.Infof("Server is shutting down")

		if err != nil {
			runError = err
			cancel()
		}
	}()

	const reasonableWaitTimeToInitializeServer = 50 * time.Millisecond
	time.Sleep(reasonableWaitTimeToInitializeServer)

	<-waitCtx.Done()

	fmt.Printf("shutting down\n")
	err := a.shutdown()
	fmt.Printf("waiting for shutting down finished\n")
	if err != nil {
		return err
	} else if runError != nil {
		return runError
	}

	a.log.Infof("Server gracefully shutdown")
	return nil
}

func (a *acceptorImpl) initializeListener() error {
	var err error

	address := fmt.Sprintf(":%d", a.port)
	a.listener, err = net.Listen("tcp", address)

	if err != nil {
		return bterr.WrapCode(err, ErrTcpInitialization)
	}

	a.log.Infof("Server will be listening at %s", address)

	return nil
}

func (a *acceptorImpl) shutdown() error {
	if !a.running.CompareAndSwap(true, false) {
		// Server is already shutting down.
		return nil
	}

	err := a.listener.Close()
	a.wg.Wait()
	return err
}

func (a *acceptorImpl) acceptLoop() error {
	defer a.wg.Done()

	running := true

	for running {
		fmt.Printf("accepting loop started\n")
		conn, err := a.listener.Accept()
		if err != nil {
			running = a.running.Load()
			if running {
				a.log.Errorf("Failed to accept connection in accept: %v", err)
			} else {
				fmt.Printf("accept loop finished: %v\n", err)
			}
		}

		if running {
			fmt.Printf("received connection, accepting\n")
			a.wg.Add(1)
			go a.acceptConnection(conn)
		} else {
			fmt.Printf("can't accept server is not running anymore\n")
		}
	}

	fmt.Printf("exiting accept loop\n")

	return nil
}

func (a *acceptorImpl) acceptConnection(conn net.Conn) {
	defer a.wg.Done()

	fmt.Printf("calling onConnect callback\n")
	err := errors.SafeRunSync(
		func() {
			a.callbacks.OnConnect(conn)
		},
	)

	if err != nil {
		a.log.Warnf("Failed to accept connection from %v: %v", conn.RemoteAddr(), err)
		conn.Close()
	}

	fmt.Printf("callback called and finished\n")
}
