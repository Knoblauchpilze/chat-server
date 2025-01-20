package tcp

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestUnit_HandleConnection_StartStop(t *testing.T) {
	client, server := newTestConnection()
	opts := ConnectionHandlerOptions{}

	close := handleConnection(client, opts)
	server.Close()
	close()
}

func TestUnit_HandleConnection_StartStop_WithTimeout(t *testing.T) {
	client, _ := newTestConnection()
	opts := ConnectionHandlerOptions{
		ReadTimeout: 100 * time.Millisecond,
	}

	close := handleConnection(client, opts)
	close()

}

func TestUnit_HandleConnection_AcceptsDataComingAfterReadTimeout(t *testing.T) {
	client, server := newTestConnection()
	var actual []byte
	var called int

	read := func(id uuid.UUID, data []byte) {
		called++
		actual = data
	}

	opts := ConnectionHandlerOptions{
		ReadTimeout: 100 * time.Millisecond,
		Callbacks: ConnectionCallbacks{
			ReadDataCallback: read,
		},
	}

	asyncWriteDataToConnectionWithDelay(t, client, 150*time.Millisecond)
	close := handleConnection(server, opts)
	// Sleep long enough to allow the write to happen and the next read
	// to be triggered
	time.Sleep(300 * time.Millisecond)
	close()

	assert.Equal(t, 1, called)
	assert.Equal(t, sampleData, actual)
}

func TestUnit_HandleConnection_CallsDisconnectCallback(t *testing.T) {
	client, server := newTestConnection()
	var called int

	disconnect := func(id uuid.UUID) {
		called++
	}

	opts := ConnectionHandlerOptions{
		ReadTimeout: 100 * time.Millisecond,
		Callbacks: ConnectionCallbacks{
			DisconnectCallback: disconnect,
		},
	}

	close := handleConnection(server, opts)
	client.Close()
	close()

	assert.Equal(t, 1, called)
}

func TestUnit_HandleConnection_CallsReadDataCallback(t *testing.T) {
	client, server := newTestConnection()
	var called int
	var actual []byte

	read := func(id uuid.UUID, data []byte) {
		called++
		actual = data
	}

	opts := ConnectionHandlerOptions{
		ReadTimeout: 100 * time.Millisecond,
		Callbacks: ConnectionCallbacks{
			ReadDataCallback: read,
		},
	}

	asyncWriteDataToConnection(t, client)
	close := handleConnection(server, opts)
	close()

	assert.Equal(t, 1, called)
	assert.Equal(t, sampleData, actual)
}
