package tcp

import (
	"net"
	"sync"
	"time"

	bterr "github.com/KnoblauchPilze/backend-toolkit/pkg/errors"
	"github.com/Knoblauchpilze/chat-server/pkg/errors"
)

type ConnectionHandlerOptions struct {
	ReadTimeout time.Duration
	Callbacks   ConnectionCallbacks
}

type ConnectionCloser func()

func handleConnection(conn net.Conn, opts ConnectionHandlerOptions) ConnectionCloser {
	connOpts := connectionOptions{
		ReadTimeout: opts.ReadTimeout,
	}
	tcpConn := newConnectionWithOptions(conn, connOpts)

	var wg sync.WaitGroup
	wg.Add(1)

	quit := make(chan interface{})
	closer := func() {
		close(quit)
		wg.Wait()
	}

	// https://github.com/venilnoronha/tcp-echo-server/blob/master/main.go#L43
	go func() {
		defer wg.Done()

		running := true

		for running {
			var timeout bool
			var err error

			readPanic := errors.SafeRun(func() {
				timeout, err = readFromConnection(tcpConn, opts.Callbacks)
			})

			if timeout {
				select {
				case <-quit:
					running = false
				default:
				}
			}

			if readPanic != nil {
				opts.Callbacks.OnPanic(readPanic)
			}

			if err != nil {
				running = false
			}
		}
	}()

	return closer
}

func readFromConnection(conn Connection, callbacks ConnectionCallbacks) (timeout bool, err error) {
	var data []byte
	data, err = conn.Read()

	if err == nil {
		callbacks.OnReadData(data)
	} else if bterr.IsErrorWithCode(err, ErrClientDisconnected) {
		callbacks.OnDisconnect()
	} else if bterr.IsErrorWithCode(err, ErrReadTimeout) {
		timeout = true
		err = nil
	} else {
		callbacks.OnReadError(err)
	}

	return
}
