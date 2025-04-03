package connection

import (
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Knoblauchpilze/chat-server/pkg/errors"
	"github.com/google/uuid"
)

type ListenerOptions struct {
	Id          uuid.UUID
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

	running atomic.Bool
	wg      sync.WaitGroup
}

func New(conn net.Conn, opts ListenerOptions) Listener {
	connOpts := WithReadTimeout(opts.ReadTimeout)

	l := &listenerImpl{
		id:        opts.Id,
		conn:      WithOptions(conn, connOpts),
		callbacks: opts.Callbacks,
	}

	return l
}

func (l *listenerImpl) Id() uuid.UUID {
	return l.id
}

func (l *listenerImpl) Start() {
	if !l.running.CompareAndSwap(false, true) {
		// Listener already running.
		return
	}

	l.wg.Add(1)

	// https://github.com/venilnoronha/tcp-echo-server/blob/master/main.go#L43
	go l.activeLoop()
}

func (l *listenerImpl) Close() {
	// Voluntarily ignoring errors: there's not much we can do about it.
	// Also closing the connection even if we did not start listening.
	// This can be called multiple times and this is okay.
	defer l.conn.Close()

	if !l.running.CompareAndSwap(true, false) {
		// Listener not started.
		return
	}

	l.wg.Wait()
}

func (l *listenerImpl) activeLoop() {
	defer l.wg.Done()

	running := true
	for running {
		var readResult connectionReadResult
		var err error

		readPanic := errors.SafeRunSync(func() {
			readResult, err = readFromConnection(l.id, l.conn, l.callbacks)
		})

		if readResult.timeout {
			running = l.running.Load()
		}

		l.conn.DiscardBytes(readResult.processed)

		if readPanic != nil {
			l.callbacks.OnReadError(l.id, readPanic)
		}

		if err != nil || readPanic != nil {
			running = false
		}
	}
}
