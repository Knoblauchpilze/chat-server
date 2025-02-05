package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnit_DefaultConfig_DefinesCorrectPort(t *testing.T) {
	config := DefaultConfig()

	assert.Equal(t, uint16(80), config.Port)
}
