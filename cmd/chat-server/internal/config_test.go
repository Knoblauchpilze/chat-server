package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnit_DefaultConfig_UsesPort49152(t *testing.T) {
	config := DefaultConfig()

	assert.Equal(t, uint16(49152), config.Port)
}

func TestUnit_DefaultConfig_DoesNotSetAnyCallbacks(t *testing.T) {
	config := DefaultConfig()

	assert.Nil(t, config.Callbacks.ConnectCallback)
	assert.Nil(t, config.Callbacks.DisconnectCallback)
	assert.Nil(t, config.Callbacks.ReadErrorCallback)
	assert.Nil(t, config.Callbacks.ReadDataCallback)
}
