package connection

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

const sampleReadTimeout = 500 * time.Millisecond

var sampleListenerOptions = ListenerOptions{
	ReadTimeout: sampleReadTimeout,
}

func TestUnit_Listener_StartStop(t *testing.T) {
	client, _ := newTestConnection()
	listener := New(client, sampleListenerOptions)

	listener.Start()
	listener.Close()
}

func TestUnit_Listener_WhenDataReceived_ExpectCallbackNotified(t *testing.T) {
	client, server := newTestConnection()

	var called int
	var actualData []byte
	opts := ListenerOptions{
		ReadTimeout: sampleReadTimeout,
		Callbacks: Callbacks{
			ReadDataCallbacks: []OnReadData{
				func(id uuid.UUID, data []byte) {
					called++
					actualData = data
				},
			},
		},
	}
	listener := New(server, opts)

	asyncWriteSampleDataToConnection(t, client)
	listener.Start()
	listener.Close()

	assert.Equal(t, 1, called)
	assert.Equal(t, sampleData, actualData)
}

func TestUnit_Listener_WhenClientDisconnects_ExpectCallbackNotified(t *testing.T) {
	client, server := newTestConnection()

	var called int
	opts := ListenerOptions{
		Callbacks: Callbacks{
			DisconnectCallbacks: []OnDisconnect{
				func(id uuid.UUID) {
					called++
				},
			},
		},
	}
	listener := New(server, opts)

	client.Close()
	listener.Start()
	listener.Close()

	assert.Equal(t, 1, called)
}

func TestUnit_Listener_WhenReadDataPanics_ExpectErrorCallbackNotified(t *testing.T) {
	client, server := newTestConnection()

	var called int
	var actualErr error
	opts := ListenerOptions{
		ReadTimeout: sampleReadTimeout,
		Callbacks: Callbacks{
			ReadDataCallbacks: []OnReadData{
				func(id uuid.UUID, data []byte) {
					panic(errSample)
				},
			},
			ReadErrorCallbacks: []OnReadError{
				func(id uuid.UUID, err error) {
					called++
					actualErr = err
				},
			},
		},
	}
	listener := New(server, opts)

	asyncWriteSampleDataToConnection(t, client)
	listener.Start()
	listener.Close()

	assert.Equal(t, 1, called)
	assert.Equal(t, errSample, actualErr)
}

func TestUnit_Listener_WhenFirstReadTimeouts_ExpectDataCanStillBeRead(t *testing.T) {
	client, server := newTestConnection()

	var called int
	var actualData []byte
	opts := ListenerOptions{
		ReadTimeout: 500 * time.Millisecond,
		Callbacks: Callbacks{
			ReadDataCallbacks: []OnReadData{
				func(id uuid.UUID, data []byte) {
					called++
					actualData = data
				},
			},
		},
	}
	listener := New(server, opts)

	// Write after a longer delay than the read timeout
	asyncWriteSampleDataToConnectionWithDelay(t, client, 750*time.Millisecond)
	listener.Start()

	// Wait long enough to close to allow the write to happen
	time.Sleep(1 * time.Second)
	listener.Close()

	assert.Equal(t, 1, called)
	assert.Equal(t, sampleData, actualData)
}

func TestUnit_Listener_WhenDataReceived_ExpectCallbackReceivesCorrectId(t *testing.T) {
	client, server := newTestConnection()

	var actualId uuid.UUID
	opts := ListenerOptions{
		ReadTimeout: sampleReadTimeout,
		Callbacks: Callbacks{
			ReadDataCallbacks: []OnReadData{
				func(id uuid.UUID, data []byte) {
					actualId = id
				},
			},
		},
	}
	listener := New(server, opts)

	asyncWriteSampleDataToConnection(t, client)
	listener.Start()
	listener.Close()

	assert.Equal(t, listener.Id(), actualId)
}

func TestUnit_Listener_WhenClientDisconnects_ExpectCallbackReceivesCorrectId(t *testing.T) {
	client, server := newTestConnection()

	var actualId uuid.UUID
	opts := ListenerOptions{
		Callbacks: Callbacks{
			DisconnectCallbacks: []OnDisconnect{
				func(id uuid.UUID) {
					actualId = id
				},
			},
		},
	}
	listener := New(server, opts)

	client.Close()
	listener.Start()
	listener.Close()

	assert.Equal(t, listener.Id(), actualId)
}

func TestUnit_Listener_WhenReadDataErrors_ExpectCallbackReceivesCorrectId(t *testing.T) {
	client, server := newTestConnection()

	var actualId uuid.UUID
	opts := ListenerOptions{
		ReadTimeout: sampleReadTimeout,
		Callbacks: Callbacks{
			ReadDataCallbacks: []OnReadData{
				func(id uuid.UUID, data []byte) {
					panic(errSample)
				},
			},
			ReadErrorCallbacks: []OnReadError{
				func(id uuid.UUID, err error) {
					actualId = id
				},
			},
		},
	}
	listener := New(server, opts)

	asyncWriteSampleDataToConnection(t, client)
	listener.Start()
	listener.Close()

	assert.Equal(t, listener.Id(), actualId)
}
