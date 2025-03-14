package connection

import (
	"bufio"
	"io"
	"net"
	"time"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/errors"
)

// This represents 1 MB
const reasonableIncompleteMessageSizeInBytes = 1024 * 1024

type connection interface {
	Read() ([]byte, error)
	Write(b []byte) (int, error)
	Close() error
}

type connectionOptions struct {
	ReadTimeout                  time.Duration
	IncompleteMessageSizeInBytes int
}

type connectionImpl struct {
	conn        net.Conn
	reader      *bufio.Reader
	readTimeout time.Duration

	incompleteMessageSizeInBytes int
	incompleteData               []byte
}

func WithReadTimeout(timeout time.Duration) connectionOptions {
	return connectionOptions{
		ReadTimeout:                  timeout,
		IncompleteMessageSizeInBytes: reasonableIncompleteMessageSizeInBytes,
	}
}

func Wrap(conn net.Conn) connection {
	return WithOptions(conn, connectionOptions{})
}

func WithOptions(conn net.Conn, options connectionOptions) connection {
	c := &connectionImpl{
		conn:                         conn,
		reader:                       bufio.NewReader(conn),
		readTimeout:                  options.ReadTimeout,
		incompleteMessageSizeInBytes: options.IncompleteMessageSizeInBytes,
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
		bytes = append(c.incompleteData, bytes...)
		c.incompleteData = []byte{}
		return bytes, nil
	}

	if err == io.EOF {
		return nil, errors.NewCode(ErrClientDisconnected)
	} else if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
		if err := c.accumulateIncompleteData(bytes); err != nil {
			return []byte{}, err
		}

		return []byte{}, errors.NewCode(ErrReadTimeout)
	}

	return nil, err
}

func (c *connectionImpl) Write(data []byte) (int, error) {
	return c.conn.Write(data)
}

func (c *connectionImpl) Close() error {
	return c.conn.Close()
}

func (c *connectionImpl) accumulateIncompleteData(data []byte) error {
	c.incompleteData = append(c.incompleteData, data...)
	if len(c.incompleteData) > c.incompleteMessageSizeInBytes {
		return errors.NewCode(ErrTooLargeIncompleteData)
	}

	return nil
}
