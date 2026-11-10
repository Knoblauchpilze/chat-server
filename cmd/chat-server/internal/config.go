package internal

import (
	"time"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/db/postgresql"
	"github.com/Knoblauchpilze/backend-toolkit/pkg/server"
)

type Configuration struct {
	Server                 server.Config
	Port                   uint16
	MessageQueueSize       int
	ClientMessageQueueSize int
	Database               postgresql.Config
}

func DefaultConfig() Configuration {
	const defaultDatabaseName = "db_chat_server"
	const defaultDatabaseUser = "chat_server_manager"

	return Configuration{
		Server: server.Config{
			BasePath:        "/v1/chats",
			ShutdownTimeout: 3 * time.Second,
		},
		Port:                   uint16(80),
		MessageQueueSize:       10,
		ClientMessageQueueSize: 2,
		Database: postgresql.NewConfigForDockerContainer(
			defaultDatabaseName,
			defaultDatabaseUser,
			"comes-from-the-environment",
		),
	}
}
