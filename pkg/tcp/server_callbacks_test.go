package tcp

import (
	"testing"

	"github.com/coder/websocket"
	"github.com/stretchr/testify/assert"
)

func TestUnit_ServerCallbacks_OnConnect_WhenUnset_ExpectNoFatalFailure(t *testing.T) {
	var callbacks ServerCallbacks

	var conn *websocket.Conn

	callback := func() {
		callbacks.OnConnect(conn)
	}
	assert.NotPanics(t, callback)
}
