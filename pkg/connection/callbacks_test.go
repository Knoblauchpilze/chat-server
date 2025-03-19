package connection

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestUnit_Callbacks_OnDisconnect_WhenUnset_ExpectNoFatalFailure(t *testing.T) {
	var callbacks Callbacks

	callback := func() {
		callbacks.OnDisconnect(sampleUuid)
	}
	assert.NotPanics(t, callback)
}
func TestUnit_Callbacks_OnDisconnect_ExpectCallbackIsCalled(t *testing.T) {
	var called int

	callbacks := Callbacks{
		DisconnectCallback: func(id uuid.UUID) {
			called++
		},
	}

	callbacks.OnDisconnect(uuid.New())

	assert.Equal(t, 1, called)
}

func TestUnit_Callbacks_OnReadError_WhenUnset_ExpectNoFatalFailure(t *testing.T) {
	var callbacks Callbacks

	callback := func() {
		callbacks.OnReadError(sampleUuid, errSample)
	}
	assert.NotPanics(t, callback)
}

func TestUnit_Callbacks_OnReadError_ExpectCallbackIsCalled(t *testing.T) {
	var called int

	callbacks := Callbacks{
		ReadErrorCallback: func(id uuid.UUID, err error) {
			called++
		},
	}

	callbacks.OnReadError(uuid.New(), errSample)

	assert.Equal(t, 1, called)
}

func TestUnit_Callbacks_OnReadData_WhenUnset_ExpectNoFatalFailure(t *testing.T) {
	var callbacks Callbacks

	callback := func() {
		callbacks.OnReadData(sampleUuid, []byte{})
	}
	assert.NotPanics(t, callback)
}

func TestUnit_Callbacks_OnReadData_ExpectCallbackIsCalled(t *testing.T) {
	var called int

	callbacks := Callbacks{
		ReadDataCallback: func(id uuid.UUID, data []byte) int {
			called++
			return 0
		},
	}

	processed := callbacks.OnReadData(uuid.New(), []byte{})

	assert.Equal(t, 1, called)
	assert.Equal(t, 0, processed)
}

func TestUnit_Callbacks_OnReadData_WhenUnset_ExpectNoBytesProcessed(t *testing.T) {
	var callbacks Callbacks

	processed := callbacks.OnReadData(sampleUuid, []byte{})

	assert.Equal(t, 0, processed)
}
