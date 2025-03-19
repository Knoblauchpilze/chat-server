package clients

import (
	"fmt"
	"net"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

var errSample = fmt.Errorf("sample error")

func TestUnit_Callbacks_OnConnect_WhenUnset_ExpectNoFatalFailure(t *testing.T) {
	var callbacks Callbacks

	var conn net.Conn
	callback := func() {
		callbacks.OnConnect(uuid.New(), conn)
	}
	assert.NotPanics(t, callback)
}

func TestUnit_Callbacks_OnConnect_WhenUnset_ExpectConnectionAccepted(t *testing.T) {
	var callbacks Callbacks

	var conn net.Conn
	actual := callbacks.OnConnect(uuid.New(), conn)

	assert.True(t, actual)
}

func TestUnit_Callbacks_OnConnect_ExpectCallbackToBeCalled(t *testing.T) {
	var called int
	callbacks := Callbacks{
		ConnectCallback: func(id uuid.UUID, conn net.Conn) bool {
			called++
			return false
		},
	}

	var conn net.Conn
	callbacks.OnConnect(uuid.New(), conn)

	assert.Equal(t, 1, called)
}

func TestUnit_Callbacks_OnConnect_ExpectCallbackValueToBeReturned(t *testing.T) {
	var conn net.Conn

	callbacks := Callbacks{
		ConnectCallback: func(id uuid.UUID, conn net.Conn) bool {
			return false
		},
	}
	actual := callbacks.OnConnect(uuid.New(), conn)
	assert.False(t, actual)

	callbacks.ConnectCallback = func(id uuid.UUID, conn net.Conn) bool {
		return true
	}
	actual = callbacks.OnConnect(uuid.New(), conn)
	assert.True(t, actual)
}

func TestUnit_Callbacks_OnDisconnect_WhenUnset_ExpectNoFatalFailure(t *testing.T) {
	var callbacks Callbacks

	callback := func() {
		callbacks.OnDisconnect(uuid.New())
	}
	assert.NotPanics(t, callback)
}

func TestUnit_Callbacks_OnDisconnect_ExpectCallbackToBeCalled(t *testing.T) {
	var called int
	callbacks := Callbacks{
		DisconnectCallback: func(id uuid.UUID) {
			called++
		},
	}

	callbacks.OnDisconnect(uuid.New())

	assert.Equal(t, 1, called)
}

func TestUnit_Callbacks_OnReadData_WhenUnset_ExpectNoFatalFailure(t *testing.T) {
	var callbacks Callbacks

	callback := func() {
		callbacks.OnReadData(uuid.New(), []byte{})
	}
	assert.NotPanics(t, callback)
}

func TestUnit_Callbacks_OnReadData_WhenUnset_ExpectConnectionStaysOpen(t *testing.T) {
	var callbacks Callbacks

	_, actual := callbacks.OnReadData(uuid.New(), []byte{})

	assert.True(t, actual)
}

func TestUnit_Callbacks_OnReadData_WhenUnset_ExpectNoBytesProcessed(t *testing.T) {
	var callbacks Callbacks

	actual, _ := callbacks.OnReadData(uuid.New(), []byte{})

	assert.Equal(t, 0, actual)
}

func TestUnit_Callbacks_OnReadData_ExpectCallbackToBeCalled(t *testing.T) {
	var called int
	callbacks := Callbacks{
		ReadDataCallback: func(id uuid.UUID, data []byte) (int, bool) {
			called++
			return 0, true
		},
	}

	callbacks.OnReadData(uuid.New(), []byte{})

	assert.Equal(t, 1, called)
}

func TestUnit_Callbacks_OnReadData_ExpectCallbackValueToBeReturned(t *testing.T) {
	callbacks := Callbacks{
		ReadDataCallback: func(id uuid.UUID, data []byte) (int, bool) {
			return 14, false
		},
	}

	processed, keepAlive := callbacks.OnReadData(uuid.New(), []byte{})

	assert.Equal(t, 14, processed)
	assert.False(t, keepAlive)
}

func TestUnit_Callbacks_OnReadError_WhenUnset_ExpectNoFatalFailure(t *testing.T) {
	var callbacks Callbacks

	callback := func() {
		callbacks.OnReadError(uuid.New(), errSample)
	}
	assert.NotPanics(t, callback)
}

func TestUnit_Callbacks_OnReadError_ExpectCallbackToBeCalled(t *testing.T) {
	var called int
	callbacks := Callbacks{
		ReadErrorCallback: func(id uuid.UUID, err error) {
			called++
		},
	}

	callbacks.OnReadError(uuid.New(), errSample)

	assert.Equal(t, 1, called)
}
