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

type ConnectionAcceptor interface {
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

func NewConnectionAcceptor(config AcceptorConfig, log logger.Logger) ConnectionAcceptor {
	a := acceptorImpl{
		port:      config.Port,
		log:       log,
		callbacks: config.Callbacks,
	}

	return &a
}

func (a *acceptorImpl) Accept() error {
	if !a.running.CompareAndSwap(false, true) {
		// Acceptor is already listening
		return bterr.NewCode(ErrAlreadyListening)
	}

	if err := a.initializeListener(); err != nil {
		fmt.Printf("err listener: %v\n", err)
		return err
	}

	a.wg.Add(1)
	return a.acceptLoop()
}

func (a *acceptorImpl) Close() error {
	if !a.running.CompareAndSwap(true, false) {
		// Server is already shutting down.
		return nil
	}

	var err error
	if a.listener != nil {
		err = a.listener.Close()
	}
	a.wg.Wait()
	return err
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
