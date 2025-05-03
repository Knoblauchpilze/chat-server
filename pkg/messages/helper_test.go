package messages

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

func insertTestUser(t *testing.T, conn db.Connection) persistence.User {
	repo := repositories.NewUserRepository(conn)

	tx, err := conn.BeginTx(context.Background())
	assert.Nil(t, err, "Actual err: %v", err)

	id := uuid.New()
	user := persistence.User{
		Id:      id,
		Name:    fmt.Sprintf("my-user-%s", uuid.NewString()),
		ApiUser: uuid.New(),
	}
	out, err := repo.Create(context.Background(), tx, user)
	tx.Close(context.Background())
	assert.Nil(t, err, "Actual err: %v", err)

	assertUserExists(t, conn, out.Id)

	return out
}

func assertUserExists(t *testing.T, dbConn db.Connection, id uuid.UUID) {
	value, err := db.QueryOne[uuid.UUID](
		context.Background(),
		dbConn,
		"SELECT id FROM chat_user WHERE id = $1",
		id,
	)
	assert.Nil(t, err, "Actual err: %v", err)
	assert.Equal(t, id, value)
}

func insertTestRoom(t *testing.T, conn db.Connection) persistence.Room {
	repo := repositories.NewRoomRepository(conn)

	id := uuid.New()
	room := persistence.Room{
		Id:   id,
		Name: fmt.Sprintf("my-room-%s", id),
	}
	out, err := repo.Create(context.Background(), room)
	assert.Nil(t, err, "Actual err: %v", err)

	assertRoomExists(t, conn, out.Id)

	return out
}

func assertRoomExists(t *testing.T, dbConn db.Connection, id uuid.UUID) {
	value, err := db.QueryOne[uuid.UUID](
		context.Background(),
		dbConn,
		"SELECT id FROM room WHERE id = $1",
		id,
	)
	assert.Nil(t, err, "Actual err: %v", err)
	assert.Equal(t, id, value)
}

func insertUserInRoom(t *testing.T, conn db.Connection, user uuid.UUID, room uuid.UUID) {
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

func assertMessageExists(
	t *testing.T,
	conn db.Connection,
	user uuid.UUID,
	room uuid.UUID,
	message string,
) {
	value, err := db.QueryOne[int](
		context.Background(),
		conn,
		`SELECT count(*)
			FROM message
			WHERE chat_user = $1
			AND room = $2
			AND message = $3`,
		user,
		room,
		message,
	)

	assert.Nil(t, err, "Actual err: %v", err)
	assert.Equal(t, 1, value)
}
