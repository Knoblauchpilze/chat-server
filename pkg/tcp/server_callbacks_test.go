package tcp

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnit_ServerCallbacks_OnConnect_WhenUnset_ExpectNoFatalFailure(t *testing.T) {
	var callbacks ServerCallbacks

	conn, _ := net.Pipe()

	callback := func() {
		callbacks.OnConnect(conn)
	}
	assert.NotPanics(t, callback)
}
