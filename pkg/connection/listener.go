package connection

import (
	"net"
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

	quit chan interface{}
	done chan bool
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
	// https://github.com/venilnoronha/tcp-echo-server/blob/master/main.go#L43
	go l.activeLoop()
}

func (l *listenerImpl) Close() {
	close(l.quit)
	<-l.done
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
