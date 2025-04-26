package connection

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

var sampleUuid = uuid.New()
var errSample = fmt.Errorf("some error")
var sampleData = []byte("hello\n")

type testServer struct {
	t    *testing.T
	conn *websocket.Conn
}

func (s *testServer) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	opts := websocket.AcceptOptions{
		OriginPatterns: []string{"localhost:*"},
	}

	var err error
	s.conn, err = websocket.Accept(rw, req, &opts)

	assert.Nil(s.t, err, "Actual err: %v", err)
}

func newTestConnection(
	t *testing.T, port uint16,
) (client *websocket.Conn, server *websocket.Conn) {
	// https://github.com/coder/websocket/blob/e4379472fe1dfe70032ecc68fec08b1b3a8fc996/internal/examples/echo/server_test.go#L18
	ts := &testServer{t: t}
	s := httptest.NewServer(ts)

	var wg sync.WaitGroup
	wg.Add(1)
	asyncConnect := func() {
		defer wg.Done()

		var dialErr error
		client, _, dialErr = websocket.Dial(
			context.Background(),
			s.URL,
			nil,
		)
		assert.Nil(t, dialErr, "Actual err: %v", dialErr)
	}

	go asyncConnect()

	wg.Wait()

	s.Close()

	server = ts.conn

	return
}

func asyncWriteSampleDataToConnection(
	t *testing.T, conn *websocket.Conn,
) *sync.WaitGroup {
	return asyncWriteSampleDataToConnectionWithDelay(t, conn, 0)
}

func asyncWriteSampleDataToConnectionWithDelay(
	t *testing.T, conn *websocket.Conn, delay time.Duration,
) *sync.WaitGroup {
	var wg sync.WaitGroup
	var err error

	wg.Add(1)
	go func() {
		defer wg.Done()

		if delay > 0 {
			time.Sleep(delay)
		}

		err = conn.Write(context.Background(), websocket.MessageBinary, sampleData)
		assert.Nil(t, err, "Actual err: %v", err)
	}()

	return &wg
}

func asyncWriteDataToConnection(
	t *testing.T, conn *websocket.Conn, data []byte,
) *sync.WaitGroup {
	var wg sync.WaitGroup
	var err error

	wg.Add(1)
	go func() {
		defer wg.Done()

		err = conn.Write(context.Background(), websocket.MessageBinary, data)
		assert.Nil(t, err, "Actual err: %v", err)
	}()

	return &wg
}

func asyncReadDataFromConnection(
	t *testing.T, conn *websocket.Conn,
) (*sync.WaitGroup, []byte) {
	var wg sync.WaitGroup

	var out []byte

	wg.Add(1)
	go func() {
		defer wg.Done()

		var err error
		var msgType websocket.MessageType
		msgType, out, err = conn.Read(context.Background())
		assert.Equal(t, websocket.MessageBinary, msgType)
		assert.Nil(t, err, "Actual err: %v", err)
	}()

	return &wg, out
}

func assertConnectionIsClosed(t *testing.T, conn *websocket.Conn) {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, _, err := conn.Read(ctx)

	assert.Equal(t, io.EOF, err)
}
