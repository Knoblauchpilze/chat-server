package internal

import (
	"context"
	"fmt"
	"io"
	"net"
	"sync"
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

func asyncRunServerAndWaitForItToBeUp(
	t *testing.T,
	s Server,
	ctx context.Context,
) (*sync.WaitGroup, *error) {
	var err error
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer func() {
			if panicErr := recover(); panicErr != nil {
				assert.Failf(t, "Server panicked", "Panic details: %v", panicErr)
			}
		}()
		err = s.Start(ctx)
	}()

	time.Sleep(reasonableWaitTimeForServerToBeUp)

	return &wg, &err
}

func asyncOpenConnectionAndCloseIt(t *testing.T, port uint16) *sync.WaitGroup {
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()

		address := fmt.Sprintf(":%d", port)
		conn, err := net.Dial("tcp", address)
		assert.Nil(t, err, "Unexpected dial error: %v", err)

		conn.Close()
	}()

	return &wg
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
