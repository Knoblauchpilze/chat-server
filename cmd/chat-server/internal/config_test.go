package internal

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestUnit_DefaultConfig_DefinesCorrectRestConfiguration(t *testing.T) {
	config := DefaultConfig()

	assert.Equal(t, uint16(80), config.Server.Port)
	assert.Equal(t, 5*time.Second, config.Server.ShutdownTimeout)
}
