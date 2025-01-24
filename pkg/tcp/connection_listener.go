package tcp

import (
	"fmt"
	"net"
	"sync"
	"time"

	bterr "github.com/Knoblauchpilze/backend-toolkit/pkg/errors"
	"github.com/Knoblauchpilze/chat-server/pkg/errors"
	"github.com/google/uuid"
)

type ConnectionListenerOptions struct {
	ReadTimeout time.Duration
	Callbacks   ConnectionCallbacks
}

type ConnectionListener interface {
	Id() uuid.UUID
	StartListening()
	Close()
}

type connectionListenerImpl struct {
	id uuid.UUID

	conn      Connection
	callbacks ConnectionCallbacks

	quit chan interface{}
	wg   sync.WaitGroup
}

// Create a new connection with an already
func newListener(conn net.Conn, opts ConnectionListenerOptions) ConnectionListener {
	connOpts := ConnectionOptions{
		ReadTimeout: opts.ReadTimeout,
	}

	l := &connectionListenerImpl{
		id:        uuid.New(),
		conn:      WithOptions(conn, connOpts),
		callbacks: opts.Callbacks,
		quit:      make(chan interface{}),
	}

	return l
}

func (l *connectionListenerImpl) Id() uuid.UUID {
	return l.id
}

func (l *connectionListenerImpl) StartListening() {
	// https://github.com/venilnoronha/tcp-echo-server/blob/master/main.go#L43
	l.wg.Add(1)
	fmt.Printf("Starting listening for %v\n", l.id)
	go l.activeLoop()
}

func (l *connectionListenerImpl) Close() {
	close(l.quit)
	fmt.Printf("Waiting to close connection %v\n", l.id)
	l.wg.Wait()
	fmt.Printf("Connection %v closed\n", l.id)
}

func (l *connectionListenerImpl) activeLoop() {
	defer l.wg.Done()

	running := true
	for running {
		var timeout bool
		var err error

		readPanic := errors.SafeRun(func() {
			timeout, err = readFromConnection(l.id, l.conn, l.callbacks)
		})

		if timeout {
			select {
			case <-l.quit:
				running = false
			default:
			}
		}

		if readPanic != nil {
			l.callbacks.OnPanic(l.id, readPanic)
		}

		if err != nil {
			running = false
		}
	}
}

func readFromConnection(id uuid.UUID, conn Connection, callbacks ConnectionCallbacks) (timeout bool, err error) {
	var data []byte

	data, err = conn.Read()

	fmt.Printf("Read returned err: %v\n", err)

	if err == nil {
		callbacks.OnReadData(id, data)
	} else if bterr.IsErrorWithCode(err, ErrClientDisconnected) {
		callbacks.OnDisconnect(id)
	} else if bterr.IsErrorWithCode(err, ErrReadTimeout) {
		timeout = true
		err = nil
	} else {
		callbacks.OnReadError(id, err)
	}

	fmt.Printf("After read timeout is: %t\n", timeout)

	return
}
