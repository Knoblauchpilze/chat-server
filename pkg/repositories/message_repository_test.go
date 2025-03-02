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
		Id:      uuid.New(),
		User:    user.Id,
		Room:    room.Id,
		Message: "hello world!",
	}

	actual, err := repo.Create(context.Background(), msg)
	assert.Nil(t, err)

	assert.True(t, eassert.EqualsIgnoringFields(actual, msg, "CreatedAt"))
	assert.True(t, actual.CreatedAt.After(beforeInsertion))
	assertMessageExists(t, conn, msg.Id)
}

func TestIT_MessageRepository_Create_WhenUserDoesNotExist_ExpectFailure(t *testing.T) {
	repo, conn := newTestMessageRepository(t)

	room := insertTestRoom(t, conn)

	msg := persistence.Message{
		Id:      uuid.New(),
		User:    uuid.New(),
		Room:    room.Id,
		Message: "hello world!",
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
		Id:      uuid.New(),
		User:    user.Id,
		Room:    uuid.New(),
		Message: "hello world!",
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
	assert.Nil(t, err)
	assert.Equal(t, id, value)
}
