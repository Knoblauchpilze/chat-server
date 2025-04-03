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

	lastSuccessfulProcessing time.Time
	incompleteDataTimeout    time.Duration

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

	l.lastSuccessfulProcessing = time.Now()

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

func (l *listenerImpl) updateLastSuccessfulRead(
	readResult connectionReadResult) error {
	if readResult.processed > 0 {
		// TODO: We should also probably reset when we receive data after not having
		// had anything to process anymore for a while. Otherwise imagine that:
		//  - we receive some data
		//  - the lastSuccessfulProcessing is set
		//  - we process it all in go
		//  - we wait long enough to pass the timeout
		//  - as there's no available data it's all good
		//  - we receive partial data
		//  - we disconnect immediately
		l.lastSuccessfulProcessing = time.Now()
	}

	if readResult.available == 0 {
		return nil
	}

	if time.Since(l.lastSuccessfulProcessing) < l.incompleteDataTimeout {
		// We processed data recently enough so we can still wait
		// for more data to come in.
		return nil
	}

	// It's been long enough since we last processed data so it's
	// safer to assume that the client is either misbehaving or
	// not responsive.
	return bterrors.NewCode(ErrIncompleteDataTimeout)
}
