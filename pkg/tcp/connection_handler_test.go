package tcp

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestUnit_HandleConnection_StartStop(t *testing.T) {
	client, server := newTestConnection()
	opts := ConnectionHandlerOptions{}

	close, err := HandleConnection(client, opts)
	server.Close()
	close()

	assert.Nil(t, err)
}

func TestUnit_HandleConnection_StartStop_WithTimeout(t *testing.T) {
	client, _ := newTestConnection()
	opts := ConnectionHandlerOptions{
		ReadTimeout: 100 * time.Millisecond,
	}

	close, err := HandleConnection(client, opts)
	close()

	assert.Nil(t, err)
}

func TestUnit_HandleConnection_AcceptsDataComingAfterReadTimeout(t *testing.T) {
	client, server := newTestConnection()
	var actual []byte
	var called int

	read := func(data []byte) error {
		called++
		actual = data
		return nil
	}

	opts := ConnectionHandlerOptions{
		ReadTimeout: 100 * time.Millisecond,
		Callbacks: ConnectionCallbacks{
			ReadDataCallback: read,
		},
	}

	asyncWriteDataToConnectionWithDelay(t, client, 150*time.Millisecond)
	close, err := HandleConnection(server, opts)
	// Sleep long enough to allow the write to happen and the next read
	// to be triggered
	time.Sleep(300 * time.Millisecond)
	close()

	assert.Nil(t, err)
	assert.Equal(t, 1, called)
	assert.Equal(t, sampleData, actual)
}

func TestUnit_HandleConnection_CallsDisconnectCallback(t *testing.T) {
	client, server := newTestConnection()
	var called int

	disconnect := func() error {
		called++
		return nil
	}

	opts := ConnectionHandlerOptions{
		ReadTimeout: 100 * time.Millisecond,
		Callbacks: ConnectionCallbacks{
			DisconnectCallback: disconnect,
		},
	}

	close, err := HandleConnection(server, opts)
	client.Close()
	close()

	assert.Nil(t, err)
	assert.Equal(t, 1, called)
}

func TestUnit_HandleConnection_CallsReadDataCallback(t *testing.T) {
	client, server := newTestConnection()
	var called int
	var actual []byte

	read := func(data []byte) error {
		called++
		actual = data
		return nil
	}

	opts := ConnectionHandlerOptions{
		ReadTimeout: 100 * time.Millisecond,
		Callbacks: ConnectionCallbacks{
			ReadDataCallback: read,
		},
	}

	asyncWriteDataToConnection(t, client)
	close, err := HandleConnection(server, opts)
	close()

	assert.Nil(t, err)
	assert.Equal(t, 1, called)
	assert.Equal(t, sampleData, actual)
}
