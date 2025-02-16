package connection

import (
	"fmt"
	"io"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

var sampleUuid = uuid.New()
var errSample = fmt.Errorf("some error")
var sampleData = []byte("hello\n")

func newTestConnection() (client net.Conn, server net.Conn) {
	return net.Pipe()
}

func asyncWriteSampleDataToConnection(t *testing.T, conn net.Conn) *sync.WaitGroup {
	return asyncWriteSampleDataToConnectionWithDelay(t, conn, 0)
}

func asyncWriteSampleDataToConnectionWithDelay(t *testing.T, conn net.Conn, delay time.Duration) *sync.WaitGroup {
	var wg sync.WaitGroup
	var err error

	wg.Add(1)
	go func() {
		defer wg.Done()

		if delay > 0 {
			time.Sleep(delay)
		}

		_, err = conn.Write(sampleData)
		assert.Nil(t, err, "Actual err: %v", err)
	}()

	return &wg
}

type readResult struct {
	data []byte
	size int
}

func asyncReadDataFromConnection(t *testing.T, conn net.Conn) (*sync.WaitGroup, *readResult) {
	var wg sync.WaitGroup

	const reasonableBufferSize = 15
	out := readResult{
		data: make([]byte, reasonableBufferSize),
	}

	wg.Add(1)
	go func() {
		defer wg.Done()

		var err error
		out.size, err = conn.Read(out.data)
		assert.Nil(t, err, "Actual err: %v", err)
	}()

	return &wg, &out
}

func assertConnectionIsClosed(t *testing.T, conn net.Conn) {
	conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))

	oneByte := make([]byte, 1)
	_, err := conn.Read(oneByte)

	// As we use pipe and not real net.Conn the returned error is this one
	// and not io.EOF.
	// TODO: Replace this with real connection.
	assert.Equal(t, io.ErrClosedPipe, err)
}
