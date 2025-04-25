package internal

import (
	"time"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/db/postgresql"
	"github.com/Knoblauchpilze/backend-toolkit/pkg/server"
	"github.com/Knoblauchpilze/chat-server/pkg/clients"
	"github.com/Knoblauchpilze/chat-server/pkg/tcp"
)

type Configuration struct {
	Server         server.Config
	ConnectTimeout time.Duration
	TcpServer      tcp.ServerConfiguration
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
		TcpServer: tcp.ServerConfiguration{
			BasePath:        "/ws",
			Port:            uint16(49152),
			ShutdownTimeout: 3 * time.Second,
		},
		Database: postgresql.NewConfigForDockerContainer(
			defaultDatabaseName,
			defaultDatabaseUser,
			"comes-from-the-environment",
		),
	}
}
