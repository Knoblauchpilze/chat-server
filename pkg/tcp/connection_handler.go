package tcp

import (
	"net"
	"sync"
	"time"

	"github.com/KnoblauchPilze/backend-toolkit/pkg/errors"
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
			timeout, err := readFromConnection(tcpConn, opts.Callbacks)

			if timeout {
				select {
				case <-quit:
					running = false
				default:
				}
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
	} else if errors.IsErrorWithCode(err, ErrClientDisconnected) {
		callbacks.OnDisconnect()
	} else if errors.IsErrorWithCode(err, ErrReadTimeout) {
		timeout = true
		err = nil
	} else {
		callbacks.OnReadError(err)
	}

	return
}
