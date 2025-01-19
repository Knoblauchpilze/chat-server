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

func HandleConnection(conn net.Conn, opts ConnectionHandlerOptions) (closer ConnectionCloser, err error) {
	connOpts := connectionOptions{
		ReadTimeout: opts.ReadTimeout,
	}
	tcpConn := newConnectionWithOptions(conn, connOpts)

	var wg sync.WaitGroup
	wg.Add(1)

	quit := make(chan interface{})
	closer = func() {
		close(quit)
		wg.Wait()
	}

	// https://github.com/venilnoronha/tcp-echo-server/blob/master/main.go#L43
	go func() {
		defer wg.Done()

		running := true
		var data []byte
		var handleErr error

		for running {
			data, err = tcpConn.Read()

			if err == nil {
				handleErr = opts.Callbacks.OnReadData(data)
			} else if errors.IsErrorWithCode(err, ErrClientDisconnected) {
				handleErr = opts.Callbacks.OnDisconnect()
			} else if errors.IsErrorWithCode(err, ErrReadTimeout) {
				select {
				case <-quit:
					running = false
				default:
				}
				err = nil
			} else {
				handleErr = opts.Callbacks.OnReadError(err)
			}

			if err != nil {
				running = false
			}
		}

		err = handleErr
	}()

	return
}
