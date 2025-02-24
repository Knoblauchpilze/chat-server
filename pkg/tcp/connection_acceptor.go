package tcp

import (
	"fmt"
	"net"
	"sync"
	"sync/atomic"

	bterr "github.com/Knoblauchpilze/backend-toolkit/pkg/errors"
	"github.com/Knoblauchpilze/backend-toolkit/pkg/logger"
	"github.com/Knoblauchpilze/chat-server/pkg/errors"
)

type connectionAcceptor interface {
	// Prompt to start accepting incoming connection. This function blocks until
	// Close is called on the same acceptor, at which point it returns any error.
	// Calling it multiple times on the same acceptor will return an error.
	Accept() error

	// Interrupts any prior blocked call to Accept and prevents accepting any new
	// connection. This call blocks until all on connect events have been processed.
	// Calling this multiple times is safe although subsequent calls will return
	// immediately.
	Close() error
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

func newConnectionAcceptor(config acceptorConfig, log logger.Logger) (connectionAcceptor, error) {
	a := acceptorImpl{
		port:      config.Port,
		log:       log,
		callbacks: config.Callbacks,
	}

	var err error
	a.listener, err = initializeListener(config.Port)

	log.Infof("Server will be listening on port %v", config.Port)

	return &a, err
}

func (a *acceptorImpl) Accept() error {
	if !a.running.CompareAndSwap(false, true) {
		// Acceptor is already listening.
		return bterr.NewCode(ErrAlreadyListening)
	}

	a.wg.Add(1)
	return a.acceptLoop()
}

func (a *acceptorImpl) Close() error {
	if !a.running.CompareAndSwap(true, false) {
		// Server is already shutting down.
		return nil
	}

	err := a.listener.Close()
	a.wg.Wait()
	return err
}

func initializeListener(port uint16) (net.Listener, error) {
	address := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", address)

	if err != nil {
		return nil, bterr.WrapCode(err, ErrTcpInitialization)
	}

	return listener, nil
}

func (a *acceptorImpl) acceptLoop() error {
	defer a.wg.Done()

	running := true

	for running {
		conn, err := a.listener.Accept()
		if err != nil {
			running = a.running.Load()
			if running {
				a.log.Errorf("Failed to accept connection in accept: %v", err)
			}
		}

		if running {
			a.wg.Add(1)
			go a.acceptConnection(conn)
		}
	}

	return nil
}

func (a *acceptorImpl) acceptConnection(conn net.Conn) {
	defer a.wg.Done()

	err := errors.SafeRunSync(
		func() {
			a.callbacks.OnConnect(conn)
		},
	)

	if err != nil {
		a.log.Warnf("Failed to accept connection from %v: %v", conn.RemoteAddr(), err)
		conn.Close()
	}
}
