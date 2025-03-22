package internal

import (
	"context"
	"fmt"
	"io"
	"net"
	"testing"
	"time"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/db"
	"github.com/Knoblauchpilze/backend-toolkit/pkg/db/postgresql"
	"github.com/stretchr/testify/assert"
)

var dbTestConfig = postgresql.NewConfigForLocalhost(
	"db_chat_server",
	"chat_server_manager",
	"manager_password",
)

const reasonableWaitTimeForServerToBeUp = 200 * time.Millisecond
const reasonableReadTimeout = 100 * time.Millisecond
const reasonableReadSizeInBytes = 1024

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
	assert.Equal(t, io.EOF, err, "Actual err: %v", err)
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

	assert.True(t, isTimeoutError(err), "Actual err: %v", err)
}

func drainConnection(t *testing.T, conn net.Conn) []byte {
	conn.SetReadDeadline(time.Now().Add(reasonableReadTimeout))

	out := make([]byte, reasonableReadSizeInBytes)
	n, err := conn.Read(out)
	if err != nil && err != io.EOF && !isTimeoutError(err) {
		assert.Nil(t, err, "Actual err: %v", err)
	}

	return out[:n]
}

func isTimeoutError(err error) bool {
	opErr, ok := err.(*net.OpError)
	if !ok {
		return false
	}

	return opErr.Timeout()
}

func newTestDbConnection(t *testing.T) db.Connection {
	conn, err := db.New(context.Background(), dbTestConfig)
	assert.Nil(t, err, "Actual err: %v", err)
	return conn
}
