package tcp

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

var errSample = fmt.Errorf("some error")

func TestUnit_Callbacks_OnDisconnect_WhenUnset_ExpectNoFatalFailure(t *testing.T) {
	var callbacks ConnectionCallbacks

	var actual error
	callback := func() {
		actual = callbacks.OnDisconnect()
	}
	assert.NotPanics(t, callback)
	assert.Nil(t, actual)
}

func TestUnit_Callbacks_OnReadError_WhenUnset_ExpectNoFatalFailure(t *testing.T) {
	var callbacks ConnectionCallbacks

	var actual error
	callback := func() {
		actual = callbacks.OnReadError(errSample)
	}
	assert.NotPanics(t, callback)
	assert.Nil(t, actual)
}

func TestUnit_Callbacks_OnReadData_WhenUnset_ExpectNoFatalFailure(t *testing.T) {
	var callbacks ConnectionCallbacks

	var actual error
	callback := func() {
		actual = callbacks.OnReadData([]byte{})
	}
	assert.NotPanics(t, callback)
	assert.Nil(t, actual)
}
