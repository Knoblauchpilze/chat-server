package connection

import (
	"fmt"
	"net"
	"sync"
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

	started       atomic.Bool
	stopRequested atomic.Bool
	wg            sync.WaitGroup
}

func New(conn net.Conn, opts ListenerOptions) Listener {
	connOpts := connectionOptions{
		ReadTimeout: opts.ReadTimeout,
	}

	l := &listenerImpl{
		id:        uuid.New(),
		conn:      WithOptions(conn, connOpts),
		callbacks: opts.Callbacks,
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

	fmt.Printf("starting listener\n")

	l.wg.Add(1)

	// https://github.com/venilnoronha/tcp-echo-server/blob/master/main.go#L43
	go l.activeLoop()
}

func (l *listenerImpl) Close() {
	if !l.stopRequested.CompareAndSwap(false, true) {
		// Stop already requested
		return
	}
	defer l.stopRequested.Store(false)

	if l.started.CompareAndSwap(true, false) {
		fmt.Printf("closing listener\n")
		l.wg.Wait()
		fmt.Printf("closing listener done\n")
	}

	// Voluntarily ignoring errors: there's not much we can do about it.
	fmt.Printf("closing connection from listener\n")
	l.conn.Close()
}

func (l *listenerImpl) activeLoop() {
	defer l.wg.Done()

	running := true
	for running {
		var timeout bool
		var err error

		readPanic := errors.SafeRunSync(func() {
			timeout, err = readFromConnection(l.id, l.conn, l.callbacks)
		})

		if timeout {
			if l.stopRequested.Load() {
				running = false
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
