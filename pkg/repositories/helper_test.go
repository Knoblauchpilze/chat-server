package repositories

import (
	"context"
	"testing"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/db"
	"github.com/Knoblauchpilze/backend-toolkit/pkg/db/postgresql"
	"github.com/stretchr/testify/require"
)

var dbTestConfig = postgresql.NewConfigForLocalhost(
	"db_chat_server",
	"chat_server_manager",
	"manager_password",
)

func newTestConnection(t *testing.T) db.Connection {
	conn, err := db.New(context.Background(), dbTestConfig)
	require.Nil(t, err)
	return conn
}
