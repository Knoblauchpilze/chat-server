package tcp

import (
	"net"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestUnit_ConnectionListener_StartStop(t *testing.T) {
	client, server := newTestConnection()
	opts := ConnectionListenerOptions{}

	listener := createAndStartListener(client, opts)
	server.Close()
	listener.Close()
}

func TestUnit_HandleConnection_StartStop_WithTimeout(t *testing.T) {
	client, _ := newTestConnection()
	opts := ConnectionListenerOptions{
		ReadTimeout: 100 * time.Millisecond,
	}

	listener := createAndStartListener(client, opts)
	listener.Close()
}

func TestUnit_HandleConnection_AcceptsDataComingAfterReadTimeout(t *testing.T) {
	client, server := newTestConnection()
	var actual []byte
	var called int

	read := func(id uuid.UUID, data []byte) {
		called++
		actual = data
	}

	opts := ConnectionListenerOptions{
		ReadTimeout: 100 * time.Millisecond,
		Callbacks: ConnectionCallbacks{
			ReadDataCallbacks: []OnReadData{read},
		},
	}

	asyncWriteDataToConnectionWithDelay(t, client, 150*time.Millisecond)
	listener := createAndStartListener(server, opts)
	// Sleep long enough to allow the write to happen and the next read
	// to be triggered
	time.Sleep(300 * time.Millisecond)
	listener.Close()

	assert.Equal(t, 1, called)
	assert.Equal(t, sampleData, actual)
}

func TestUnit_HandleConnection_CallsDisconnectCallback(t *testing.T) {
	client, server := newTestConnection()
	var called int

	disconnect := func(id uuid.UUID) {
		called++
	}

	opts := ConnectionListenerOptions{
		ReadTimeout: 100 * time.Millisecond,
		Callbacks: ConnectionCallbacks{
			DisconnectCallbacks: []OnDisconnect{disconnect},
		},
	}

	listener := createAndStartListener(server, opts)
	client.Close()
	listener.Close()

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

	opts := ConnectionListenerOptions{
		ReadTimeout: 100 * time.Millisecond,
		Callbacks: ConnectionCallbacks{
			ReadDataCallbacks: []OnReadData{read},
		},
	}

	asyncWriteDataToConnection(t, client)
	listener := createAndStartListener(server, opts)
	listener.Close()

	assert.Equal(t, 1, called)
	assert.Equal(t, sampleData, actual)
}

func createAndStartListener(conn net.Conn, opts ConnectionListenerOptions) ConnectionListener {
	listener := newListener(conn, opts)
	listener.StartListening()
	return listener
}
