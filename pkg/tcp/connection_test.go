package tcp

import (
	"net"
	"sync"
	"testing"
	"time"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/errors"
	"github.com/stretchr/testify/assert"
)

// https://stackoverflow.com/questions/30688685/how-does-one-test-net-conn-in-unit-tests-in-golang

var sampleData = []byte("hello\n")

func TestUnit_Connection_Read(t *testing.T) {
	client, server := newTestConnection()
	conn := Wrap(server)

	wg, writeErr := asyncWriteDataToConnection(t, client)
	actual, err := conn.Read()
	wg.Wait()

	assert.Nil(t, *writeErr)
	assert.Nil(t, err)
	assert.Equal(t, sampleData, actual)
}

func TestUnit_Connection_ReadWithTimeout_WhenNoDataWritten_ReturnsNoData(t *testing.T) {
	_, server := newTestConnection()
	opts := connectionOptions{
		ReadTimeout: 500 * time.Millisecond,
	}
	conn := newConnectionWithOptions(server, opts)

	actual, err := conn.Read()

	assert.True(t, errors.IsErrorWithCode(err, ErrReadTimeout))
	assert.Equal(t, []byte{}, actual)
}

func TestUnit_Connection_ReadWithTimeout(t *testing.T) {
	client, server := newTestConnection()
	opts := connectionOptions{
		// 2 reads will be over the delay we set for the client connection
		ReadTimeout: 700 * time.Millisecond,
	}
	conn := newConnectionWithOptions(server, opts)

	wg, writeErr := asyncWriteDataToConnectionWithDelay(t, client, 1*time.Second)
	firstRead, err := conn.Read()
	assert.True(t, errors.IsErrorWithCode(err, ErrReadTimeout))

	secondRead, err := conn.Read()
	assert.Nil(t, err)

	wg.Wait()

	assert.Nil(t, *writeErr)
	assert.Equal(t, []byte{}, firstRead)
	assert.Equal(t, sampleData, secondRead)
}

func TestUnit_Connection_Read_WhenDisconnect_ReturnsExplicitError(t *testing.T) {
	client, server := newTestConnection()
	conn := Wrap(server)

	err := client.Close()
	assert.Nil(t, err)
	actual, err := conn.Read()

	assert.True(t, errors.IsErrorWithCode(err, ErrClientDisconnected), "Actual err: %v", err)
	assert.Nil(t, actual)
}

func TestUnit_Connection_Write(t *testing.T) {
	client, server := newTestConnection()
	conn := Wrap(server)

	wg, actual := asyncReadDataFromConnection(client)
	sizeWritten, err := conn.Write(sampleData)
	wg.Wait()

	assert.Nil(t, err)
	assert.Nil(t, actual.err)
	assert.Equal(t, sizeWritten, actual.size)
	assert.Equal(t, sampleData, actual.data[:actual.size])
}

func TestUnit_Connection_Write_WhenDisconnect_ReturnsExplicitError(t *testing.T) {
	client, server := newTestConnection()
	conn := Wrap(server)

	err := client.Close()
	assert.Nil(t, err)
	actual, err := conn.Write(sampleData)

	assert.True(t, errors.IsErrorWithCode(err, ErrClientDisconnected), "Actual err: %v", err)
	assert.Equal(t, 0, actual)
}

func newTestConnection() (client net.Conn, server net.Conn) {
	return net.Pipe()
}

func asyncWriteDataToConnection(t *testing.T, conn net.Conn) (*sync.WaitGroup, *error) {
	return asyncWriteDataToConnectionWithDelay(t, conn, 0)
}

func asyncWriteDataToConnectionWithDelay(t *testing.T, conn net.Conn, delay time.Duration) (*sync.WaitGroup, *error) {
	var wg sync.WaitGroup
	var err error

	wg.Add(1)
	go func() {
		defer wg.Done()

		if delay > 0 {
			time.Sleep(delay)
		}

		_, err = conn.Write(sampleData)
		assert.Nil(t, err)
	}()

	return &wg, &err
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
