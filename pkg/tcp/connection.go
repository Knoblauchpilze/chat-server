package tcp

import (
	"bufio"
	"io"
	"net"
	"time"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/errors"
)

type Connection interface {
	Read() ([]byte, error)
	Write(b []byte) (int, error)
	Close() error
}

type ConnectionOptions struct {
	ReadTimeout time.Duration
}

type connectionImpl struct {
	conn        net.Conn
	reader      *bufio.Reader
	readTimeout time.Duration
}

func Wrap(conn net.Conn) Connection {
	return WithOptions(conn, ConnectionOptions{})
}

func WithOptions(conn net.Conn, options ConnectionOptions) Connection {
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
		return bytes, errors.NewCode(ErrReadTimeout)
	}

	return nil, err
}

func (c *connectionImpl) Write(data []byte) (int, error) {
	n, err := c.conn.Write(data)

	if err == io.ErrClosedPipe {
		return n, errors.NewCode(ErrClientDisconnected)
	}

	return n, err
}

func (c *connectionImpl) Close() error {
	return c.conn.Close()
}
