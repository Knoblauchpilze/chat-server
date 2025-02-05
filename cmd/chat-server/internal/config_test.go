package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnit_DefaultConfig_DefinesCorrectRestConfiguration(t *testing.T) {
	config := DefaultConfig()

	assert.Equal(t, uint16(80), config.Server.Port)
}
