package tcp

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

const reasonableWaitTimeForAcceptorToBeUp = 200 * time.Millisecond

var errSample = fmt.Errorf("some error")
var sampleData = []byte("hello\n")

func newTestConnection(
	t *testing.T,
	port uint16,
) (client net.Conn, server net.Conn) {
	address := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", address)
	assert.Nil(t, err, "Actual err: %v", err)

	var wg sync.WaitGroup
	wg.Add(1)
	asyncConnect := func() {
		defer wg.Done()

		// Wait for the listener to be started in the main thread.
		time.Sleep(50 * time.Millisecond)

		client, err = net.Dial("tcp", address)
		assert.Nil(t, err, "Actual err: %v", err)
	}

	go asyncConnect()

	server, err = listener.Accept()
	assert.Nil(t, err, "Actual err: %v", err)

	wg.Wait()

	listener.Close()

	return
}

func asyncCancelContext(delay time.Duration, cancel context.CancelFunc) {
	go func() {
		time.Sleep(delay)
		cancel()
	}()
}

func asyncRunAcceptorAndWaitForItToBeUp(
	t *testing.T,
	ca ConnectionAcceptor,
	ctx context.Context,
) *sync.WaitGroup {
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer func() {
			if panicErr := recover(); panicErr != nil {
				assert.Failf(t, "Server panicked", "Panic details: %v", panicErr)
			}
		}()
		err := ca.Listen(ctx)
		assert.Nil(t, err, "Actual err: %v", err)
	}()

	time.Sleep(reasonableWaitTimeForAcceptorToBeUp)

	return &wg
}

func asyncOpenConnectionAndCloseIt(t *testing.T, port uint16) *sync.WaitGroup {
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()

		address := fmt.Sprintf(":%d", port)
		fmt.Printf("opening %d\n", port)
		conn, err := net.Dial("tcp", address)
		assert.Nil(t, err, "Unexpected dial error: %v", err)

		conn.Close()

		fmt.Printf("finished opening and closing connection\n")
	}()

	return &wg
}

func assertConnectionIsClosed(t *testing.T, conn net.Conn) {
	conn.SetReadDeadline(time.Now().Add(50 * time.Millisecond))

	oneByte := make([]byte, 1)
	_, err := conn.Read(oneByte)

	assert.Equal(t, io.EOF, err, "Actual err: %v", err)
}

func assertConnectionIsStillOpen(t *testing.T, conn net.Conn) {
	conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))

	oneByte := make([]byte, 1)
	_, err := conn.Read(oneByte)

	opErr, ok := err.(*net.OpError)
	assert.True(t, ok, "Error is not an OpError: %v", err)
	if ok {
		assert.True(t, opErr.Timeout(), "Actual err: %v", err)
	}
}
