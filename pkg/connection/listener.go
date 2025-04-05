package connection

import (
	"net"
	"sync"
	"sync/atomic"
	"time"

	bterrors "github.com/Knoblauchpilze/backend-toolkit/pkg/errors"
	"github.com/Knoblauchpilze/chat-server/pkg/errors"
	"github.com/google/uuid"
)

type ListenerOptions struct {
	Id uuid.UUID

	// Defines how long the listener will wait for each read operation
	// on the connection to complete. This is used to periodically interrupt
	// the listening process to allow the server to properly close connections
	// and gracefully shutdown.
	// If no data is available within the allocated time and the server is
	// still running, the listener will continue to listen for new data.
	ReadTimeout time.Duration

	// Defines how long the listener will wait in case some data is available
	// but can't be processed. It is expected that some read operations might
	// yield incomplete data. We expect the server to promptly receive the rest
	// of the data in case a legitimate client is connected.
	// If the server does not receive the rest of the data within the allocated
	// time we will terminate the connection in order to prevent resource hogging
	// where a client would open many connections to the server and send as much
	// data as possible without going over the limit and then wait indefinitely.
	IncompleteDataTimeout time.Duration

	Callbacks Callbacks
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

	availableBytesCount       int
	dataProcessingWindowStart time.Time
	incompleteDataTimeout     time.Duration

	running atomic.Bool
	wg      sync.WaitGroup
}

func New(conn net.Conn, opts ListenerOptions) Listener {
	connOpts := WithReadTimeout(opts.ReadTimeout)

	l := &listenerImpl{
		id:        opts.Id,
		conn:      WithOptions(conn, connOpts),
		callbacks: opts.Callbacks,

		incompleteDataTimeout: opts.IncompleteDataTimeout,
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

	l.dataProcessingWindowStart = time.Now()

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

		if err := l.updateLastSuccessfulRead(readResult); err != nil && running {
			l.callbacks.OnReadError(l.id, err)
		} else if readPanic != nil {
			l.callbacks.OnReadError(l.id, readPanic)
		}

		if err != nil || readPanic != nil {
			running = false
		}
	}
}

func (l *listenerImpl) updateLastSuccessfulRead(readResult connectionReadResult) error {
	if readResult.processed > 0 {
		l.dataProcessingWindowStart = time.Now()
	}

	if readResult.available != 0 && l.availableBytesCount == 0 {
		l.dataProcessingWindowStart = time.Now()
	}

	l.availableBytesCount = readResult.available

	if l.availableBytesCount == 0 {
		return nil
	}

	durationOfCurrentProcessingWindow := time.Since(l.dataProcessingWindowStart)
	if durationOfCurrentProcessingWindow < l.incompleteDataTimeout {
		// The processing window is still within what we consider valid for a
		// client to send data. Let's wait a bit longer
		return nil
	}

	// The current processing window lasts for too long already. It can either
	// be that the client is not responsive (network issue) or that it is
	// misbehaving. In both cases, we want to terminate the connection to not
	// risk resource hogging.
	return bterrors.NewCode(ErrIncompleteDataTimeout)
}
