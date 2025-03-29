package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/db"
	"github.com/Knoblauchpilze/backend-toolkit/pkg/db/pgx"
	"github.com/Knoblauchpilze/backend-toolkit/pkg/errors"
	"github.com/Knoblauchpilze/chat-server/pkg/persistence"
	eassert "github.com/Knoblauchpilze/easy-assert/assert"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestIT_MessageRepository_Create(t *testing.T) {
	repo, conn := newTestMessageRepository(t)

	room := insertTestRoom(t, conn)
	user := insertTestUser(t, conn)

	beforeInsertion := time.Now()

	msg := persistence.Message{
		Id:       uuid.New(),
		ChatUser: user.Id,
		Room:     room.Id,
		Message:  "hello world!",
	}

	actual, err := repo.Create(context.Background(), msg)
	assert.Nil(t, err, "Actual err: %v", err)

	assert.True(t, eassert.EqualsIgnoringFields(actual, msg, "CreatedAt"))
	assert.True(t, actual.CreatedAt.After(beforeInsertion))
	assertMessageExists(t, conn, msg.Id)
}

func TestIT_MessageRepository_Create_WhenUserDoesNotExist_ExpectFailure(t *testing.T) {
	repo, conn := newTestMessageRepository(t)

	room := insertTestRoom(t, conn)

	msg := persistence.Message{
		Id:       uuid.New(),
		ChatUser: uuid.New(),
		Room:     room.Id,
		Message:  "hello world!",
	}

	_, err := repo.Create(context.Background(), msg)

	assert.True(
		t,
		errors.IsErrorWithCode(err, pgx.ForeignKeyValidation),
		"Actual err: %v",
		err,
	)
	assertUserDoesNotExist(t, conn, msg.Id)
}

func TestIT_MessageRepository_Create_WhenRoomDoesNotExist_ExpectFailure(t *testing.T) {
	repo, conn := newTestMessageRepository(t)

	user := insertTestUser(t, conn)

	msg := persistence.Message{
		Id:       uuid.New(),
		ChatUser: user.Id,
		Room:     uuid.New(),
		Message:  "hello world!",
	}

	_, err := repo.Create(context.Background(), msg)

	assert.True(
		t,
		errors.IsErrorWithCode(err, pgx.ForeignKeyValidation),
		"Actual err: %v",
		err,
	)
	assertUserDoesNotExist(t, conn, msg.Id)
}

func TestIT_MessageRepository_ListForRoom(t *testing.T) {
	repo, conn := newTestMessageRepository(t)
	room1 := insertTestRoom(t, conn)
	room2 := insertTestRoom(t, conn)
	user1 := insertTestUser(t, conn)
	registerUserInRoom(t, conn, user1.Id, room1.Id)
	user2 := insertTestUser(t, conn)
	registerUserInRoom(t, conn, user2.Id, room2.Id)

	msg1 := insertTestMessage(t, conn, user1.Id, room1.Id)
	insertTestMessage(t, conn, user2.Id, room2.Id)

	actual, err := repo.ListForRoom(context.Background(), room1.Id)
	assert.Nil(t, err, "Actual err: %v", err)

	assert.Len(t, actual, 1)
	assert.True(t, eassert.EqualsIgnoringFields(actual[0], msg1))
}

func TestIT_MessageRepository_ListForRoom_WhenNoMessageAvailable_ReturnsEmptySlice(t *testing.T) {
	repo, conn := newTestMessageRepository(t)
	room := insertTestRoom(t, conn)

	actual, err := repo.ListForRoom(context.Background(), room.Id)
	assert.Nil(t, err, "Actual err: %v", err)

	assert.Equal(t, []persistence.Message{}, actual)
}

func newTestMessageRepository(t *testing.T) (MessageRepository, db.Connection) {
	conn := newTestConnection(t)
	return NewMessageRepository(conn), conn
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

func insertTestMessage(
	t *testing.T,
	conn db.Connection,
	user uuid.UUID,
	room uuid.UUID,
) persistence.Message {
	someTime := time.Date(2025, 3, 29, 11, 42, 43, 0, time.UTC)

	msg := persistence.Message{
		Id:        uuid.New(),
		ChatUser:  user,
		Room:      room,
		Message:   "my-name-" + uuid.NewString(),
		CreatedAt: someTime,
	}

	n, err := conn.Exec(
		context.Background(),
		`INSERT INTO
			message (id, chat_user, room, message, created_at)
			VALUES ($1, $2, $3, $4, $5)`,
		msg.Id,
		msg.ChatUser,
		msg.Room,
		msg.Message,
		msg.CreatedAt,
	)
	assert.Nil(t, err, "Actual err: %v", err)
	assert.Equal(t, int64(1), n, "Actual err: %v", err)

	return msg
}
