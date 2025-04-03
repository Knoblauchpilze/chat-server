package connection

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/errors"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

const sampleReadTimeout = 150 * time.Millisecond

var sampleListenerOptions = ListenerOptions{
	ReadTimeout: sampleReadTimeout,
}

func TestUnit_Listener_UsesProvidedId(t *testing.T) {
	_, server := newTestConnection(t, 1214)
	opts := ListenerOptions{
		Id: sampleUuid,
	}

	listener := New(server, opts)

	assert.Equal(t, sampleUuid, listener.Id())
}

func TestUnit_Listener_StartStop(t *testing.T) {
	_, server := newTestConnection(t, 1200)
	listener := New(server, sampleListenerOptions)

	listener.Start()
	listener.Close()
}

func TestUnit_Listener_WhenStopping_ExpectAttachedConnectionToBeClosed(t *testing.T) {
	client, server := newTestConnection(t, 1201)
	listener := New(server, sampleListenerOptions)

	listener.Start()
	listener.Close()

	assertConnectionIsClosed(t, client)
}

func TestUnit_Listener_WhenNotStartedAndStopping_ExpectAttachedConnectionToBeClosed(t *testing.T) {
	client, server := newTestConnection(t, 1202)
	listener := New(server, sampleListenerOptions)

	listener.Close()

	assertConnectionIsClosed(t, client)
}

func TestUnit_Listener_WhenStartedMultipleTimes_ShouldStillStopAfterOneClose(t *testing.T) {
	client, server := newTestConnection(t, 1203)
	listener := New(server, sampleListenerOptions)

	listener.Start()
	listener.Start()
	listener.Close()

	assertConnectionIsClosed(t, client)
}

func TestUnit_Listener_WhenClosingMultiple_ShouldSucceed(t *testing.T) {
	_, server := newTestConnection(t, 1204)
	listener := New(server, sampleListenerOptions)

	listener.Start()

	listener.Close()
	listener.Close()
}

func TestUnit_Listener_WhenDataReceived_ExpectCallbackNotified(t *testing.T) {
	client, server := newTestConnection(t, 1205)

	var called int
	opts := ListenerOptions{
		ReadTimeout: sampleReadTimeout,
		Callbacks: Callbacks{
			ReadDataCallback: func(id uuid.UUID, data []byte) int {
				called++
				return len(data)
			},
		},
	}
	listener := New(server, opts)

	wg := asyncWriteSampleDataToConnection(t, client)
	listener.Start()

	wg.Wait()
	listener.Close()

	assert.Equal(t, 1, called)
}

func TestUnit_Listener_WhenDataReceived_ExpectCallbackReceivesCorrectId(t *testing.T) {
	client, server := newTestConnection(t, 1206)

	var actualId uuid.UUID
	opts := ListenerOptions{
		ReadTimeout: sampleReadTimeout,
		Callbacks: Callbacks{
			ReadDataCallback: func(id uuid.UUID, data []byte) int {
				actualId = id
				return 0
			},
		},
	}
	listener := New(server, opts)

	wg := asyncWriteSampleDataToConnection(t, client)
	listener.Start()

	wg.Wait()
	listener.Close()

	assert.Equal(t, listener.Id(), actualId)
}

func TestUnit_Listener_WhenDataReceived_ExpectCallbackReceivesCorrectData(t *testing.T) {
	client, server := newTestConnection(t, 1207)

	var actualData []byte
	opts := ListenerOptions{
		ReadTimeout: sampleReadTimeout,
		Callbacks: Callbacks{
			ReadDataCallback: func(id uuid.UUID, data []byte) int {
				actualData = data
				return 0
			},
		},
	}
	listener := New(server, opts)

	wg := asyncWriteSampleDataToConnection(t, client)
	listener.Start()

	wg.Wait()
	listener.Close()

	assert.Equal(t, sampleData, actualData)
}

func TestUnit_Listener_WhenDataReceivedAndProcessed_ExpectDataToBeDiscarded(t *testing.T) {
	client, server := newTestConnection(t, 1213)

	var receivedData [][]byte
	var called int

	opts := ListenerOptions{
		ReadTimeout: 50 * time.Millisecond,
		Callbacks: Callbacks{
			ReadDataCallback: func(id uuid.UUID, data []byte) int {
				called++
				receivedData = append(receivedData, data)
				return 5
			},
		},
	}
	listener := New(server, opts)

	data := []byte("123456789")
	_, err := client.Write(data)
	assert.Nil(t, err, "Actual err: %v", err)

	// Start the listener and wait for it to process the data. We
	// need to wait for a bit over twice the timeout to pass.
	listener.Start()
	time.Sleep(120 * time.Millisecond)

	listener.Close()

	assert.Equal(t, 2, called)
	assert.Equal(t, [][]byte{[]byte("123456789"), []byte("6789")}, receivedData)
}

func TestUnit_Listener_WhenClientDisconnects_ExpectCallbackNotified(t *testing.T) {
	client, server := newTestConnection(t, 1208)

	var called int
	opts := ListenerOptions{
		Callbacks: Callbacks{
			DisconnectCallback: func(id uuid.UUID) {
				called++
			},
		},
	}
	listener := New(server, opts)

	client.Close()
	listener.Start()
	listener.Close()

	assert.Equal(t, 1, called)
}

func TestUnit_Listener_WhenClientDisconnects_ExpectCallbackReceivesCorrectId(t *testing.T) {
	client, server := newTestConnection(t, 1209)

	var actualId uuid.UUID
	opts := ListenerOptions{
		Callbacks: Callbacks{
			DisconnectCallback: func(id uuid.UUID) {
				actualId = id
			},
		},
	}
	listener := New(server, opts)

	client.Close()
	listener.Start()
	listener.Close()

	assert.Equal(t, listener.Id(), actualId)
}

func TestUnit_Listener_WhenReadDataPanics_ExpectErrorCallbackNotified(t *testing.T) {
	client, server := newTestConnection(t, 1210)

	var called int
	var actualErr error
	opts := ListenerOptions{
		ReadTimeout: sampleReadTimeout,
		Callbacks: Callbacks{
			ReadDataCallback: func(id uuid.UUID, data []byte) int {
				panic(errSample)
			},
			ReadErrorCallback: func(id uuid.UUID, err error) {
				called++
				actualErr = err
			},
		},
	}
	listener := New(server, opts)

	wg := asyncWriteSampleDataToConnection(t, client)
	listener.Start()

	wg.Wait()
	listener.Close()

	assert.Equal(t, 1, called)
	assert.Equal(t, errSample, actualErr)
}

func TestUnit_Listener_WhenReadDataPanics_ExpectCallbackReceivesCorrectId(t *testing.T) {
	client, server := newTestConnection(t, 1211)

	var actualId uuid.UUID
	opts := ListenerOptions{
		ReadTimeout: sampleReadTimeout,
		Callbacks: Callbacks{
			ReadDataCallback: func(id uuid.UUID, data []byte) int {
				panic(errSample)
			},
			ReadErrorCallback: func(id uuid.UUID, err error) {
				actualId = id
			},
		},
	}
	listener := New(server, opts)

	wg := asyncWriteSampleDataToConnection(t, client)
	listener.Start()

	wg.Wait()
	listener.Close()

	assert.Equal(t, listener.Id(), actualId)
}

func TestUnit_Listener_WhenFirstReadTimeouts_ExpectDataCanStillBeRead(t *testing.T) {
	client, server := newTestConnection(t, 1212)

	var called int
	var actualData []byte
	opts := ListenerOptions{
		ReadTimeout: 100 * time.Millisecond,
		Callbacks: Callbacks{
			ReadDataCallback: func(id uuid.UUID, data []byte) int {
				called++
				actualData = data
				return len(data)
			},
		},
	}
	listener := New(server, opts)

	// Write after a longer delay than the read timeout.
	wg := asyncWriteSampleDataToConnectionWithDelay(t, client, 150*time.Millisecond)
	listener.Start()

	wg.Wait()
	listener.Close()

	assert.Equal(t, 1, called)
	assert.Equal(t, sampleData, actualData)
}

func TestUnit_Listener_WhenIncompleteDataReceived_IfNoDataComesLater_ExpectOnReadErrorNotified(t *testing.T) {
	client, server := newTestConnection(t, 1213)

	var readErr error
	var called atomic.Bool
	opts := ListenerOptions{
		ReadTimeout:           50 * time.Millisecond,
		IncompleteDataTimeout: 100 * time.Millisecond,
		Callbacks: Callbacks{
			ReadDataCallback: func(id uuid.UUID, data []byte) int {
				if called.CompareAndSwap(false, true) {
					return 1
				}
				return 0
			},
			ReadErrorCallback: func(id uuid.UUID, err error) {
				readErr = err
			},
		},
	}
	listener := New(server, opts)

	n, err := client.Write(sampleData)
	assert.Equal(t, len(sampleData), n)
	assert.Nil(t, err, "Actual err: %v", err)

	listener.Start()

	// Wait long enough for the timeout to be reached.
	time.Sleep(200 * time.Millisecond)

	listener.Close()

	assert.True(
		t,
		errors.IsErrorWithCode(readErr, ErrIncompleteDataTimeout),
		"Actual err: %v",
		readErr,
	)
}
