package internal

import (
	"testing"
	"time"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/server"
	"github.com/stretchr/testify/assert"
)

func TestUnit_DefaultConfig_DefinesCorrectRestConfiguration(t *testing.T) {
	config := DefaultConfig()

	expectedConf := server.Config{
		BasePath:        "/v1/chats",
		Port:            uint16(80),
		ShutdownTimeout: 3 * time.Second,
	}
	assert.Equal(t, expectedConf, config.Server)
}

func TestUnit_DefaultConfig_DefinesReasonableMessageQueueSize(t *testing.T) {
	config := DefaultConfig()

	assert.Equal(t, 10, config.MessageQueueSize)
}

func TestUnit_DefaultConfig_DefinesReasonableClientMessageQueueSize(t *testing.T) {
	config := DefaultConfig()

	assert.Equal(t, 2, config.ClientMessageQueueSize)
}

func TestUnit_DefaultConfig_SetsExpectedDbConnection(t *testing.T) {
	config := DefaultConfig()

	assert.Equal(t, "172.17.0.1", config.Database.Host)
	assert.Equal(t, "db_chat_server", config.Database.Database)
	assert.Equal(t, "chat_server_manager", config.Database.User)
}

func TestUnit_DefaultConfig_DoesNotSetDbPassword(t *testing.T) {
	config := DefaultConfig()

	assert.Equal(t, "comes-from-the-environment", config.Database.Password)
}
