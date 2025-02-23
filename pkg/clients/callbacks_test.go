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

func TestUnit_Callbacks_OnDisonnect_WhenUnset_ExpectNoFatalFailure(t *testing.T) {
	var callbacks Callbacks

	callback := func() {
		callbacks.OnDisconnect(uuid.New())
	}
	assert.NotPanics(t, callback)
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

	actual := callbacks.OnReadData(uuid.New(), []byte{})

	assert.True(t, actual)
}

func TestUnit_Callbacks_OnReadError_WhenUnset_ExpectNoFatalFailure(t *testing.T) {
	var callbacks Callbacks

	callback := func() {
		callbacks.OnReadError(uuid.New(), errSample)
	}
	assert.NotPanics(t, callback)
}
