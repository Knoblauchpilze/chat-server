package service

import (
	"fmt"
	"io"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const reasonableWaitTimeForServiceToBeUp = 50 * time.Millisecond
const reasonableReadTimeout = 100 * time.Millisecond
const reasonableReadSizeInBytes = 1024

func asyncRunChatService(
	t *testing.T, service ChatService,
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
		service.Start()
	}()

	time.Sleep(reasonableWaitTimeForServiceToBeUp)

	return &wg
}

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

		var dialErr error
		client, dialErr = net.Dial("tcp", address)
		assert.Nil(t, dialErr, "Actual err: %v", dialErr)
	}

	go asyncConnect()

	server, err = listener.Accept()
	assert.Nil(t, err, "Actual err: %v", err)

	wg.Wait()

	listener.Close()

	return
}

func readFromConnection(t *testing.T, conn net.Conn) []byte {
	conn.SetReadDeadline(time.Now().Add(reasonableReadTimeout))

	out := make([]byte, reasonableReadSizeInBytes)
	n, err := conn.Read(out)
	assert.Nil(t, err, "Actual err: %v", err)

	return out[:n]
}

func drainConnection(t *testing.T, conn net.Conn) []byte {
	conn.SetReadDeadline(time.Now().Add(reasonableReadTimeout))

	out := make([]byte, reasonableReadSizeInBytes)
	n, err := conn.Read(out)
	if err != nil && err != io.EOF {
		assert.Nil(t, err, "Actual err: %v", err)
	}

	return out[:n]
}

func assertNoDataReceived(t *testing.T, conn net.Conn) {
	conn.SetReadDeadline(time.Now().Add(reasonableReadTimeout))

	oneByte := make([]byte, 1)
	_, err := conn.Read(oneByte)

	opErr, ok := err.(*net.OpError)
	assert.True(t, ok)
	if ok {
		assert.True(t, opErr.Timeout())
	}
}
