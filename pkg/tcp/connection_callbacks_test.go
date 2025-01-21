package tcp

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

var errSample = fmt.Errorf("some error")
var sampleUuid = uuid.New()

func TestUnit_ConnectionCallbacks_OnDisconnect_WhenUnset_ExpectNoFatalFailure(t *testing.T) {
	var callbacks ConnectionCallbacks

	callback := func() {
		callbacks.OnDisconnect(sampleUuid)
	}
	assert.NotPanics(t, callback)
}

func TestUnit_ConnectionCallbacks_OnDisconnect_CallsAllCallbacks(t *testing.T) {
	var called1, called2 int

	callbacks := ConnectionCallbacks{
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

func TestUnit_ConnectionCallbacks_OnReadError_WhenUnset_ExpectNoFatalFailure(t *testing.T) {
	var callbacks ConnectionCallbacks

	callback := func() {
		callbacks.OnReadError(sampleUuid, errSample)
	}
	assert.NotPanics(t, callback)
}

func TestUnit_ConnectionCallbacks_OnReadError_CallsAllCallbacks(t *testing.T) {
	var called1, called2 int

	callbacks := ConnectionCallbacks{
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

func TestUnit_ConnectionCallbacks_OnPanic_WhenUnset_ExpectNoFatalFailure(t *testing.T) {
	var callbacks ConnectionCallbacks

	callback := func() {
		callbacks.OnPanic(sampleUuid, errSample)
	}
	assert.NotPanics(t, callback)
}

func TestUnit_ConnectionCallbacks_OnPanic_CallsAllCallbacks(t *testing.T) {
	var called1, called2 int

	callbacks := ConnectionCallbacks{
		PanicCallbacks: []OnPanic{
			func(id uuid.UUID, err error) {
				called1++
			},
			func(id uuid.UUID, err error) {
				called2++
			},
		},
	}

	callbacks.OnPanic(uuid.New(), errSample)

	assert.Equal(t, 1, called1)
	assert.Equal(t, 1, called2)
}

func TestUnit_ConnectionCallbacks_OnReadData_WhenUnset_ExpectNoFatalFailure(t *testing.T) {
	var callbacks ConnectionCallbacks

	callback := func() {
		callbacks.OnReadData(sampleUuid, []byte{})
	}
	assert.NotPanics(t, callback)
}

func TestUnit_ConnectionCallbacks_OnReadData_CallsAllCallbacks(t *testing.T) {
	var called1, called2 int

	callbacks := ConnectionCallbacks{
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
