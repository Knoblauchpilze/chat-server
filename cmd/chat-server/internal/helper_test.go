package internal

import (
	"context"
	"fmt"
	"io"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const reasonableWaitTimeForServerToBeUp = 200 * time.Millisecond

var errSample = fmt.Errorf("some error")
var sampleData = []byte("hello\n")

func asyncCancelContext(delay time.Duration, cancel context.CancelFunc) {
	go func() {
		time.Sleep(delay)
		cancel()
	}()
}

func assertConnectionIsClosed(t *testing.T, conn net.Conn) {
	conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))

	oneByte := make([]byte, 1)
	_, err := conn.Read(oneByte)

	assert.Equal(t, io.EOF, err)
}

func assertConnectionIsStillOpen(t *testing.T, conn net.Conn) {
	conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))

	oneByte := make([]byte, 1)
	_, err := conn.Read(oneByte)

	opErr, ok := err.(*net.OpError)
	assert.True(t, ok)
	if ok {
		assert.True(t, opErr.Timeout())
	}
}
