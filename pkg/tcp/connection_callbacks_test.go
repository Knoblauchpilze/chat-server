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

func TestUnit_ConnectionCallbacks_OnReadError_WhenUnset_ExpectNoFatalFailure(t *testing.T) {
	var callbacks ConnectionCallbacks

	callback := func() {
		callbacks.OnReadError(sampleUuid, errSample)
	}
	assert.NotPanics(t, callback)
}

func TestUnit_ConnectionCallbacks_OnPanic_WhenUnset_ExpectNoFatalFailure(t *testing.T) {
	var callbacks ConnectionCallbacks

	callback := func() {
		callbacks.OnPanic(sampleUuid, errSample)
	}
	assert.NotPanics(t, callback)
}

func TestUnit_ConnectionCallbacks_OnReadData_WhenUnset_ExpectNoFatalFailure(t *testing.T) {
	var callbacks ConnectionCallbacks

	callback := func() {
		callbacks.OnReadData(sampleUuid, []byte{})
	}
	assert.NotPanics(t, callback)
}
