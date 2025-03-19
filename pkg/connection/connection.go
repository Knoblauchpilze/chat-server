package connection

import (
	"io"
	"net"
	"time"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/errors"
)

// This is used as the maximum size of a message that can be received
// by our system. If a message is larger than this, we will not wait
// to process it and terminate the connection.
const maxMessageSizeInBytes = 1024
const readSizeInBytes = 512

type connection interface {
	Read() ([]byte, error)
	DiscardBytes(n int)
	Write(b []byte) (int, error)
	Close() error
}

type connectionOptions struct {
	ReadTimeout               time.Duration
	MaximumMessageSizeInBytes int
}

type connectionImpl struct {
	conn        net.Conn
	readTimeout time.Duration

	maximumMessageSizeInBytes int

	// The data waiting to be processed. This is used to accumulate
	// incoming data until some of it is processed. We only allow
	// to accumulate up to `maximumMessageSizeInBytes` bytes.
	data []byte

	// A temporary buffer to read incoming data to. This avoids
	// allocating a new buffer every time we read data.
	tmp []byte
}

func WithReadTimeout(timeout time.Duration) connectionOptions {
	return connectionOptions{
		ReadTimeout:               timeout,
		MaximumMessageSizeInBytes: maxMessageSizeInBytes,
	}
}

// Wrap wraps a connection with a default maximum message size which
// should be suited for most cases.
func Wrap(conn net.Conn) connection {
	opts := connectionOptions{
		MaximumMessageSizeInBytes: maxMessageSizeInBytes,
	}
	return WithOptions(conn, opts)
}

func WithOptions(conn net.Conn, options connectionOptions) connection {
	c := &connectionImpl{
		conn:                      conn,
		readTimeout:               options.ReadTimeout,
		maximumMessageSizeInBytes: options.MaximumMessageSizeInBytes,
		data:                      make([]byte, 0),
		tmp:                       make([]byte, readSizeInBytes),
	}

	return c
}

func (c *connectionImpl) Read() ([]byte, error) {
	if c.readTimeout > 0 {
		timeout := time.Now().Add(c.readTimeout)
		c.conn.SetReadDeadline(timeout)
	}

	received, readErr := c.conn.Read(c.tmp)

	if err := c.accumulateIncomingData(c.tmp[:received]); err != nil {
		return []byte{}, err
	}

	if readErr == io.EOF {
		return nil, errors.NewCode(ErrClientDisconnected)
	} else if opErr, ok := readErr.(*net.OpError); ok && opErr.Timeout() {
		return c.data, errors.NewCode(ErrReadTimeout)
	}

	return c.data, readErr
}

func (c *connectionImpl) DiscardBytes(n int) {
	if n >= len(c.data) {
		c.data = []byte{}
	} else {
		c.data = c.data[n:]
	}
}

func (c *connectionImpl) Write(data []byte) (int, error) {
	return c.conn.Write(data)
}

func (c *connectionImpl) Close() error {
	return c.conn.Close()
}

func (c *connectionImpl) accumulateIncomingData(data []byte) error {
	c.data = append(c.data, data...)
	if len(c.data) > c.maximumMessageSizeInBytes {
		return errors.NewCode(ErrTooLargeIncompleteData)
	}

	return nil
}
