package internal

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestUnit_DefaultConfig_UsesPort80(t *testing.T) {
	config := DefaultConfig()

	assert.Equal(t, uint16(80), config.Port)
}
func TestUnit_DefaultConfig_Defines3sReadTimeout(t *testing.T) {
	config := DefaultConfig()

	assert.Equal(t, 3*time.Second, config.ReadTimeout)
}
