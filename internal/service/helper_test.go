package service

import (
	"context"
	"fmt"
	"testing"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/db"
	"github.com/Knoblauchpilze/backend-toolkit/pkg/db/postgresql"
	"github.com/Knoblauchpilze/chat-server/pkg/persistence"
	"github.com/Knoblauchpilze/chat-server/pkg/repositories"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

var dbTestConfig = postgresql.NewConfigForLocalhost(
	"db_chat_server",
	"chat_server_manager",
	"manager_password",
)

func newTestDbConnection(t *testing.T) db.Connection {
	conn, err := db.New(context.Background(), dbTestConfig)
	assert.Nil(t, err, "Actual err: %v", err)
	return conn
}

func getRoomId(t *testing.T, conn db.Connection, name string) uuid.UUID {
	sqlQuery := `SELECT id FROM room WHERE name = $1`

	id, err := db.QueryOne[uuid.UUID](
		context.Background(),
		conn,
		sqlQuery,
		name,
	)
	assert.Nil(t, err, "Actual err: %v", err)
	return id
}

func registerUserInRoom(t *testing.T, conn db.Connection, user uuid.UUID, room uuid.UUID) {
	sqlQuery := `INSERT INTO room_user (room, chat_user) VALUES ($1, $2)`

	count, err := conn.Exec(
		context.Background(),
		sqlQuery,
		room,
		user,
	)
	assert.Nil(t, err, "Actual err: %v", err)
	assert.Equal(t, int64(1), count)
}

func registerUserByNameInRoom(t *testing.T, conn db.Connection, user string, room uuid.UUID) {
	sqlQuery := `
		INSERT INTO
			room_user (chat_user, room)
		SELECT
			id,
			$2
		FROM
			chat_user
		WHERE
			name = $1`

	count, err := conn.Exec(
		context.Background(),
		sqlQuery,
		user,
		room,
	)
	assert.Nil(t, err, "Actual err: %v", err)
	assert.Equal(t, int64(1), count)
}

func assertUserRegisteredInRoom(
	t *testing.T, conn db.Connection, user uuid.UUID, room string,
) {
	value, err := db.QueryOne[int](
		context.Background(),
		conn,
		`SELECT
			COUNT(*)
		FROM
			room_user AS ru
			LEFT JOIN room AS r ON ru.room = r.id
		WHERE
			ru.chat_user = $1
			AND r.name = $2`,
		user,
		room,
	)
	assert.Nil(t, err, "Actual err: %v", err)
	assert.Equal(t, 1, value)
}

func assertUserNameRegisteredInRoom(
	t *testing.T, conn db.Connection, user string, room uuid.UUID,
) {
	value, err := db.QueryOne[int](
		context.Background(),
		conn,
		`SELECT
			COUNT(*)
		FROM
			room_user AS ru
			LEFT JOIN chat_user AS cu ON cu.id = ru.chat_user
		WHERE
			cu.name = $1
			AND ru.room = $2`,
		user,
		room,
	)
	assert.Nil(t, err, "Actual err: %v", err)
	assert.Equal(t, 1, value)
}

func assertUserNotRegisteredInRoom(
	t *testing.T, conn db.Connection, user uuid.UUID, room string,
) {
	t.Helper()

	value, err := db.QueryOne[int](
		context.Background(),
		conn,
		`SELECT
			COUNT(*)
		FROM
			room_user AS ru
			LEFT JOIN room AS r ON ru.room = r.id
		WHERE
			ru.chat_user = $1
			AND r.name = $2`,
		user,
		room,
	)
	assert.Nil(t, err, "Actual err: %v", err)
	assert.Equal(t, 0, value)
}

func insertTestMessage(
	t *testing.T,
	conn db.Connection,
	user uuid.UUID,
	room uuid.UUID,
) persistence.Message {
	repo := repositories.NewMessageRepository(conn)

	id := uuid.New()
	msg := persistence.Message{
		Id:       id,
		ChatUser: user,
		Room:     room,
		Message:  fmt.Sprintf("my-message-%s", id),
	}
	out, err := repo.Create(context.Background(), msg)
	assert.Nil(t, err, "Actual err: %v", err)

	assertMessageExists(t, conn, out.Id)

	return out
}

func assertMessageExists(t *testing.T, conn db.Connection, id uuid.UUID) {
	value, err := db.QueryOne[uuid.UUID](
		context.Background(),
		conn,
		"SELECT id FROM message WHERE id = $1",
		id,
	)
	assert.Nil(t, err, "Actual err: %v", err)
	assert.Equal(t, id, value)
}

func assertMessageDoesNotExist(t *testing.T, conn db.Connection, id uuid.UUID) {
	value, err := db.QueryOne[int](
		context.Background(),
		conn,
		"SELECT COUNT(id) FROM message WHERE id = $1",
		id,
	)
	assert.Nil(t, err, "Actual err: %v", err)
	assert.Zero(t, value)
}

func assertMessageOwner(t *testing.T, conn db.Connection, msg uuid.UUID, user string) {
	t.Helper()

	value, err := db.QueryOne[string](
		context.Background(),
		conn,
		`SELECT
			cu.name
		FROM
			message AS m
			LEFT JOIN chat_user AS cu ON cu.id = m.chat_user
		WHERE
			m.id = $1`,
		msg,
	)
	assert.Nil(t, err, "Actual err: %v", err)
	assert.Equal(t, user, value)
}
