package connection

import (
	"fmt"
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
	err  error
}

func asyncReadDataFromConnection(conn net.Conn) (*sync.WaitGroup, *readResult) {
	var wg sync.WaitGroup

	const reasonableBufferSize = 15
	out := readResult{
		data: make([]byte, reasonableBufferSize),
	}

	wg.Add(1)
	go func() {
		defer wg.Done()

		out.size, out.err = conn.Read(out.data)
	}()

	return &wg, &out
}
