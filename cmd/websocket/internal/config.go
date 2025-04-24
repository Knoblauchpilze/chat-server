package internal

import (
	"time"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/db/postgresql"
	"github.com/Knoblauchpilze/backend-toolkit/pkg/server"
	"github.com/Knoblauchpilze/chat-server/pkg/clients"
)

type Configuration struct {
	Server         server.Config
	ConnectTimeout time.Duration
	TcpPort        uint16
	Database       postgresql.Config
	Callbacks      clients.Callbacks
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
		ConnectTimeout: 1 * time.Second,
		// https://serverfault.com/questions/11806/which-ports-to-use-on-a-self-written-tcp-server
		TcpPort: uint16(49152),
		Database: postgresql.NewConfigForDockerContainer(
			defaultDatabaseName,
			defaultDatabaseUser,
			"comes-from-the-environment",
		),
	}
}
