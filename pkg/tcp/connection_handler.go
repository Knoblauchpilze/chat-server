package tcp

import (
	"bufio"
	"fmt"
	"io"
	"net"

	"github.com/KnoblauchPilze/backend-toolkit/pkg/logger"
)

type ConnectionHandler interface {
	Handle(conn net.Conn) error
}

type connHandlerImpl struct {
	log logger.Logger
}

func newHandler(log logger.Logger) ConnectionHandler {
	return &connHandlerImpl{
		log: log,
	}
}

func (h *connHandlerImpl) Handle(conn net.Conn) error {
	// https://github.com/venilnoronha/tcp-echo-server/blob/master/main.go#L43
	h.log.Infof("Received connection")

	reader := bufio.NewReader(conn)
	for {
		// read client request data
		bytes, err := reader.ReadBytes(byte('\n'))
		if err != nil {
			if err == io.EOF {
				h.log.Warnf("Client disconnected")
				return nil
			}

			h.log.Errorf("Failed to read data, err: %v", err)
			return err
		}

		h.log.Infof("Received data (%d byte(s)): %s", len(bytes), bytes)
		out := fmt.Sprintf("Received %d byte(s)", len(bytes))
		conn.Write([]byte(out))
	}
}
