package clients

import (
	"fmt"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const reasonableReadTimeout = 100 * time.Millisecond
const reasonableReadSizeInBytes = 1024

func newTestConnection(t *testing.T, port uint16) (client net.Conn, server net.Conn) {
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
