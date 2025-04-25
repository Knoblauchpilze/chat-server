package tcp

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	bterr "github.com/Knoblauchpilze/backend-toolkit/pkg/errors"
	"github.com/Knoblauchpilze/backend-toolkit/pkg/logger"
	"github.com/Knoblauchpilze/chat-server/pkg/errors"
	"github.com/coder/websocket"
)

type websocketAcceptor interface {
	// Prompt to start accepting incoming connections. This function blocks until
	// Close is called on the same acceptor, at which point it returns any error.
	// Calling it multiple times on the same acceptor will return an error.
	Accept() error

	// Interrupts any prior blocked call to Accept and prevents accepting any new
	// connection. This call blocks until all on connect events have been processed.
	// Calling this multiple times is safe although subsequent calls will return
	// immediately.
	Close() error
}

type websocketAcceptorImpl struct {
	log logger.Logger

	shutdownTimeout time.Duration

	callbacks ServerCallbacks

	listener http.Server
	running  atomic.Bool
	wg       sync.WaitGroup
}

func NewWebsocketAcceptor(config acceptorConfig, log logger.Logger) websocketAcceptor {
	a := websocketAcceptorImpl{
		log:             log,
		shutdownTimeout: config.ShutdownTimeout,
		callbacks:       config.Callbacks,
	}

	a.initializeWebSocketListener(config.BasePath, config.Port)

	log.Infof("Server will be listening on port %v", config.Port)

	return &a
}

func (a *websocketAcceptorImpl) Accept() error {
	if !a.running.CompareAndSwap(false, true) {
		// Acceptor is already listening.
		return bterr.NewCode(ErrAlreadyListening)
	}

	a.wg.Add(1)
	return a.acceptLoop()
}

func (a *websocketAcceptorImpl) Close() error {
	if !a.running.CompareAndSwap(true, false) {
		// Server is already shutting down.
		return nil
	}

	err := a.shutdown()

	a.wg.Wait()

	return err
}

func (a *websocketAcceptorImpl) initializeWebSocketListener(
	basePath string, port uint16,
) {
	// https://shijuvar.medium.com/building-rest-apis-with-go-1-22-http-servemux-2115f242f02b
	mux := http.NewServeMux()

	fmt.Printf("handling base path: %s\n", basePath)

	mux.HandleFunc(basePath, func(rw http.ResponseWriter, req *http.Request) {
		a.handleConnectionRequest(req, rw)
	})

	address := fmt.Sprintf(":%d", port)

	a.listener = http.Server{
		Addr:    address,
		Handler: mux,
	}
}

func (a *websocketAcceptorImpl) acceptLoop() error {
	defer a.wg.Done()
	return a.listener.ListenAndServe()
}

func (a *websocketAcceptorImpl) shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), a.shutdownTimeout)
	defer cancel()
	return a.listener.Shutdown(ctx)
}

func (a *websocketAcceptorImpl) handleConnectionRequest(req *http.Request, rw http.ResponseWriter) {
	opts := websocket.AcceptOptions{
		OriginPatterns: []string{"localhost:*"},
	}

	conn, err := websocket.Accept(rw, req, &opts)
	if err != nil {
		// https://github.com/coder/websocket/blob/master/example_test.go#L21
		a.log.Errorf("Failed to upgrade connection: %v", err)
		return
	}

	if a.running.Load() {
		a.wg.Add(1)
		go a.acceptConnection(conn)
	}
}

// TODO: This should be put as part of the http handler
func (a *websocketAcceptorImpl) acceptConnection(conn *websocket.Conn) {
	defer a.wg.Done()

	err := errors.SafeRunSync(
		func() {
			a.callbacks.OnConnect(conn)
		},
	)

	if err != nil {
		a.log.Warnf("Failed to accept connection: %v", err)
		// TODO: Double check if this code makes sense
		conn.Close(websocket.StatusNormalClosure, "denied")
	}
}
