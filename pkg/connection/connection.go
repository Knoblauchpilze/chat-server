package connection

import (
	"bufio"
	"io"
	"net"
	"time"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/errors"
)

type connection interface {
	Read() ([]byte, error)
	Write(b []byte) (int, error)
	Close() error
}

type connectionOptions struct {
	ReadTimeout time.Duration
}

type connectionImpl struct {
	conn        net.Conn
	reader      *bufio.Reader
	readTimeout time.Duration
}

func Wrap(conn net.Conn) connection {
	return WithOptions(conn, connectionOptions{})
}

func WithOptions(conn net.Conn, options connectionOptions) connection {
	c := &connectionImpl{
		conn:        conn,
		reader:      bufio.NewReader(conn),
		readTimeout: options.ReadTimeout,
	}

	return c
}

func (c *connectionImpl) Read() ([]byte, error) {
	if c.readTimeout > 0 {
		timeout := time.Now().Add(c.readTimeout)
		c.conn.SetReadDeadline(timeout)
	}

	bytes, err := c.reader.ReadBytes(byte('\n'))
	if err == nil {
		return bytes, nil
	}

	if err == io.EOF {
		return nil, errors.NewCode(ErrClientDisconnected)
	} else if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
		// TODO: We could return nil. This is important in case we receive data
		// that is not terminated by a newline. In such cases the next read will
		// not return any data but it might be that combining both would be
		// helpful.
		// Maybe we could keep an internal buffer of data that was read but not
		// returned to the caller.
		// Additionally we should have some protection to not just keep adding
		// more buffered data if the client is not sending any newline.
		return bytes, errors.NewCode(ErrReadTimeout)
	}

	return nil, err
}

func (c *connectionImpl) Write(data []byte) (int, error) {
	return c.conn.Write(data)
}

func (c *connectionImpl) Close() error {
	return c.conn.Close()
}
