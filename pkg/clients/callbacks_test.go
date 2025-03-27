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
		callbacks.OnConnect(conn)
	}
	assert.NotPanics(t, callback)
}

func TestUnit_Callbacks_OnConnect_WhenUnset_ExpectConnectionRefused(t *testing.T) {
	var callbacks Callbacks

	var conn net.Conn
	actual, _ := callbacks.OnConnect(conn)

	assert.False(t, actual)
}

func TestUnit_Callbacks_OnConnect_WhenUnset_ExpectNilUuid(t *testing.T) {
	var callbacks Callbacks

	var conn net.Conn
	_, actual := callbacks.OnConnect(conn)

	assert.Equal(t, uuid.Nil, actual)
}

func TestUnit_Callbacks_OnConnect_ExpectCallbackToBeCalled(t *testing.T) {
	var called int
	callbacks := Callbacks{
		ConnectCallback: func(conn net.Conn) (bool, uuid.UUID) {
			called++
			return false, uuid.Nil
		},
	}

	var conn net.Conn
	callbacks.OnConnect(conn)

	assert.Equal(t, 1, called)
}

func TestUnit_Callbacks_OnConnect_ExpectCallbackValueToBeReturned(t *testing.T) {
	var conn net.Conn

	callbacks := Callbacks{
		ConnectCallback: func(conn net.Conn) (bool, uuid.UUID) {
			return false, uuid.Nil
		},
	}
	actual, id := callbacks.OnConnect(conn)
	assert.False(t, actual)
	assert.Equal(t, uuid.Nil, id)

	expected := uuid.New()
	callbacks.ConnectCallback = func(conn net.Conn) (bool, uuid.UUID) {
		return true, expected
	}
	actual, id = callbacks.OnConnect(conn)
	assert.True(t, actual)
	assert.Equal(t, expected, id)
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
