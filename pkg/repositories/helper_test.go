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

func assertUserRegisteredInRoom(
	t *testing.T, conn db.Connection, user uuid.UUID, room uuid.UUID,
) {
	value, err := db.QueryOne[int](
		context.Background(),
		conn,
		"SELECT COUNT(*) FROM room_user WHERE chat_user = $1 AND room = $2",
		user,
		room,
	)
	assert.Nil(t, err, "Actual err: %v", err)
	assert.Equal(t, 1, value)
}

func assertUserNotRegisteredInRoom(
	t *testing.T, conn db.Connection, user uuid.UUID, room uuid.UUID,
) {
	t.Helper()

	value, err := db.QueryOne[int](
		context.Background(),
		conn,
		"SELECT COUNT(*) FROM room_user WHERE chat_user = $1 AND room = $2",
		user,
		room,
	)
	assert.Nil(t, err, "Actual err: %v", err)
	assert.Equal(t, 0, value)
}
