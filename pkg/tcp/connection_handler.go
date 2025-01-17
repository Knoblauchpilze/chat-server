package tcp

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/KnoblauchPilze/backend-toolkit/pkg/errors"
	"github.com/KnoblauchPilze/backend-toolkit/pkg/logger"
)

type ConnectionHandler interface {
	Handle(conn net.Conn) error
	Close()
}

type connHandlerImpl struct {
	log         logger.Logger
	readTimeout time.Duration

	quit chan interface{}
	wg   sync.WaitGroup
}

func newHandler(readTimeout time.Duration, log logger.Logger) ConnectionHandler {
	return &connHandlerImpl{
		log:         log,
		readTimeout: readTimeout,

		quit: make(chan interface{}),
	}
}

func (h *connHandlerImpl) Handle(conn net.Conn) error {
	// https://github.com/venilnoronha/tcp-echo-server/blob/master/main.go#L43
	h.log.Debugf("Received connection from %v", conn.RemoteAddr())

	defer h.wg.Done()
	h.wg.Add(1)

	reader := bufio.NewReader(conn)

	var readErr error

	running := true
	for running {
		timeout := time.Now().Add(h.readTimeout)
		conn.SetReadDeadline(timeout)

		data, err := h.readData(reader)
		if err != nil {
			if errors.IsErrorWithCode(err, ErrClientDisconnected) {
				h.log.Debugf("Client disconnected")
				running = false
			} else if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
				select {
				case <-h.quit:
					running = false
				default:
				}
			} else {
				h.log.Errorf("Failed to read data, err: %v", err)
				readErr = err
			}
		}

		if running && err == nil {
			h.log.Infof("Received data (%d byte(s)): %s", len(data), data)
			h.sendData(conn, data)
		}
	}

	return readErr
}

func (h *connHandlerImpl) Close() {
	close(h.quit)
	h.wg.Wait()
}

func (h *connHandlerImpl) readData(reader *bufio.Reader) ([]byte, error) {
	bytes, err := reader.ReadBytes(byte('\n'))
	if err == nil {
		return bytes, nil
	}

	if err == io.EOF {
		return nil, errors.NewCode(ErrClientDisconnected)
	}
	return nil, err
}

func (h *connHandlerImpl) sendData(conn net.Conn, data []byte) {
	out := fmt.Sprintf("Received %d byte(s)", len(data))

	_, err := conn.Write([]byte(out))
	if err != nil {
		h.log.Errorf("Failed to send %d byte(s), err: %v", len(out), err)
	}
}
