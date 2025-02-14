package connection

import (
	"fmt"
	"net"
	"sync/atomic"
	"time"

	"github.com/Knoblauchpilze/chat-server/pkg/errors"
	"github.com/google/uuid"
)

type ListenerOptions struct {
	ReadTimeout time.Duration
	Callbacks   Callbacks
}

type Listener interface {
	Id() uuid.UUID
	Start()
	Close()
}

type listenerImpl struct {
	id uuid.UUID

	conn      connection
	callbacks Callbacks

	started atomic.Bool
	quit    chan interface{}
	done    chan bool
}

func New(conn net.Conn, opts ListenerOptions) Listener {
	connOpts := connectionOptions{
		ReadTimeout: opts.ReadTimeout,
	}

	l := &listenerImpl{
		id:        uuid.New(),
		conn:      WithOptions(conn, connOpts),
		callbacks: opts.Callbacks,
		quit:      make(chan interface{}),
		done:      make(chan bool),
	}

	return l
}

func (l *listenerImpl) Id() uuid.UUID {
	return l.id
}

func (l *listenerImpl) Start() {
	if !l.started.CompareAndSwap(false, true) {
		// Listener already started
		return
	}

	// https://github.com/venilnoronha/tcp-echo-server/blob/master/main.go#L43
	go l.activeLoop()
}

func (l *listenerImpl) Close() {
	close(l.quit)
	fmt.Printf("waiting for listener close\n")
	if l.started.Load() {
		<-l.done
	}
	// Voluntarily ignoring errors: there's not much we can do about it.
	fmt.Printf("closing connection\n")
	l.conn.Close()
}

func (l *listenerImpl) activeLoop() {
	defer func() {
		l.done <- true
	}()

	running := true
	for running {
		var timeout bool
		var err error

		readPanic := errors.SafeRunSync(func() {
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
			l.callbacks.OnReadError(l.id, readPanic)
		}

		if err != nil {
			running = false
		}
	}
}
