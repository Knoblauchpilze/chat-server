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

func TestUnit_Callbacks_OnDisconnect_CallsAllCallbacks(t *testing.T) {
	var called1, called2 int

	callbacks := Callbacks{
		DisconnectCallbacks: []OnDisconnect{
			func(id uuid.UUID) {
				called1++
			},
			func(id uuid.UUID) {
				called2++
			},
		},
	}

	callbacks.OnDisconnect(uuid.New())

	assert.Equal(t, 1, called1)
	assert.Equal(t, 1, called2)
}

func TestUnit_Callbacks_OnReadError_WhenUnset_ExpectNoFatalFailure(t *testing.T) {
	var callbacks Callbacks

	callback := func() {
		callbacks.OnReadError(sampleUuid, errSample)
	}
	assert.NotPanics(t, callback)
}

func TestUnit_ConnectionCallbacks_OnReadError_CallsAllCallbacks(t *testing.T) {
	var called1, called2 int

	callbacks := Callbacks{
		ReadErrorCallbacks: []OnReadError{
			func(id uuid.UUID, err error) {
				called1++
			},
			func(id uuid.UUID, err error) {
				called2++
			},
		},
	}

	callbacks.OnReadError(uuid.New(), errSample)

	assert.Equal(t, 1, called1)
	assert.Equal(t, 1, called2)
}

func TestUnit_Callbacks_OnReadData_WhenUnset_ExpectNoFatalFailure(t *testing.T) {
	var callbacks Callbacks

	callback := func() {
		callbacks.OnReadData(sampleUuid, []byte{})
	}
	assert.NotPanics(t, callback)
}

func TestUnit_ConnectionCallbacks_OnReadData_CallsAllCallbacks(t *testing.T) {
	var called1, called2 int

	callbacks := Callbacks{
		ReadDataCallbacks: []OnReadData{
			func(id uuid.UUID, data []byte) {
				called1++
			},
			func(id uuid.UUID, data []byte) {
				called2++
			},
		},
	}

	callbacks.OnReadData(uuid.New(), []byte{})

	assert.Equal(t, 1, called1)
	assert.Equal(t, 1, called2)
}
