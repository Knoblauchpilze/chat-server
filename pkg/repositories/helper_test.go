package repositories

import (
	"context"
	"testing"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/db"
	"github.com/Knoblauchpilze/backend-toolkit/pkg/db/postgresql"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

var dbTestConfig = postgresql.NewConfigForLocalhost(
	"db_chat_server",
	"chat_server_manager",
	"manager_password",
)

func newTestConnection(t *testing.T) db.Connection {
	conn, err := db.New(context.Background(), dbTestConfig)
	assert.Nil(t, err, "Actual err: %v", err)
	return conn
}

func registerUserInRoom(
	t *testing.T, conn db.Connection, user uuid.UUID, room uuid.UUID,
) {
	_, err := conn.Exec(
		context.Background(),
		`INSERT INTO room_user (room, chat_user) VALUES ($1, $2)`,
		room,
		user,
	)
	assert.Nil(t, err, "Actual err: %v", err)
}
