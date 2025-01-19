package tcp

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestUnit_HandleConnection_StartStop(t *testing.T) {
	client, server := newTestConnection()
	opts := ConnectionHandlerOptions{}

	close := HandleConnection(client, opts)
	server.Close()
	close()
}

func TestUnit_HandleConnection_StartStop_WithTimeout(t *testing.T) {
	client, _ := newTestConnection()
	opts := ConnectionHandlerOptions{
		ReadTimeout: 100 * time.Millisecond,
	}

	close := HandleConnection(client, opts)
	close()

}

func TestUnit_HandleConnection_AcceptsDataComingAfterReadTimeout(t *testing.T) {
	client, server := newTestConnection()
	var actual []byte
	var called int

	read := func(data []byte) {
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
	close := HandleConnection(server, opts)
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

	disconnect := func() {
		called++
	}

	opts := ConnectionHandlerOptions{
		ReadTimeout: 100 * time.Millisecond,
		Callbacks: ConnectionCallbacks{
			DisconnectCallback: disconnect,
		},
	}

	close := HandleConnection(server, opts)
	client.Close()
	close()

	assert.Equal(t, 1, called)
}

func TestUnit_HandleConnection_CallsReadDataCallback(t *testing.T) {
	client, server := newTestConnection()
	var called int
	var actual []byte

	read := func(data []byte) {
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
	close := HandleConnection(server, opts)
	close()

	assert.Equal(t, 1, called)
	assert.Equal(t, sampleData, actual)
}
