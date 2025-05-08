package internal

import (
	"time"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/db/postgresql"
	"github.com/Knoblauchpilze/backend-toolkit/pkg/server"
)

type Configuration struct {
	Server           server.Config
	MessageQueueSize int
	Database         postgresql.Config
}

func DefaultConfig() Configuration {
	const defaultDatabaseName = "db_chat_server"
	const defaultDatabaseUser = "chat_server_manager"

	return Configuration{
		Server: server.Config{
			BasePath:        "/v1/chats",
			Port:            uint16(80),
			ShutdownTimeout: 3 * time.Second,
		},
		MessageQueueSize: 10,
		Database: postgresql.NewConfigForDockerContainer(
			defaultDatabaseName,
			defaultDatabaseUser,
			"comes-from-the-environment",
		),
	}
}
