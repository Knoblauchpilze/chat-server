package tcp

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

var errSample = fmt.Errorf("some error")

func TestUnit_Callbacks_OnDisconnect_WhenUnset_ExpectNoFatalFailure(t *testing.T) {
	var callbacks ConnectionCallbacks

	callback := func() {
		callbacks.OnDisconnect()
	}
	assert.NotPanics(t, callback)
}

func TestUnit_Callbacks_OnReadError_WhenUnset_ExpectNoFatalFailure(t *testing.T) {
	var callbacks ConnectionCallbacks

	callback := func() {
		callbacks.OnReadError(errSample)
	}
	assert.NotPanics(t, callback)
}

func TestUnit_Callbacks_OnReadData_WhenUnset_ExpectNoFatalFailure(t *testing.T) {
	var callbacks ConnectionCallbacks

	callback := func() {
		callbacks.OnReadData([]byte{})
	}
	assert.NotPanics(t, callback)
}
