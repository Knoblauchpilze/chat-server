package repositories

import (
	"context"
	"fmt"
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
	defer conn.Close(context.Background())
	beforeInsertion := time.Now()

	room := insertTestRoom(t, conn)
	user := insertTestUser(t, conn)
	registerUserInRoom(t, conn, user.Id, room.Id)

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

func TestIT_MessageRepository_Create_WhenUserNotRegisteredInRoom_ExpectFailure(t *testing.T) {
	repo, conn := newTestMessageRepository(t)
	defer conn.Close(context.Background())
	room := insertTestRoom(t, conn)
	user := insertTestUser(t, conn)

	msg := persistence.Message{
		Id:       uuid.New(),
		ChatUser: user.Id,
		Room:     room.Id,
		Message:  "hello world!",
	}

	_, err := repo.Create(context.Background(), msg)
	assert.True(
		t,
		errors.IsErrorWithCode(err, ErrUserNotRegisteredInRoom),
		"Actual err: %v",
		err,
	)
	assertMessageDoesNotExist(t, conn, msg.Id)
}

func TestIT_MessageRepository_Create_WhenUserDoesNotExist_ExpectFailure(t *testing.T) {
	repo, conn := newTestMessageRepository(t)
	defer conn.Close(context.Background())

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
		errors.IsErrorWithCode(err, ErrUserNotRegisteredInRoom),
		"Actual err: %v",
		err,
	)
	assertUserDoesNotExist(t, conn, msg.Id)
}

func TestIT_MessageRepository_Create_WhenRoomDoesNotExist_ExpectFailure(t *testing.T) {
	repo, conn := newTestMessageRepository(t)
	defer conn.Close(context.Background())

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
		errors.IsErrorWithCode(err, ErrUserNotRegisteredInRoom),
		"Actual err: %v",
		err,
	)
	assertUserDoesNotExist(t, conn, msg.Id)
}

func TestIT_MessageRepository_ListForRoom(t *testing.T) {
	repo, conn := newTestMessageRepository(t)
	defer conn.Close(context.Background())
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

	expected := []persistence.Message{msg1}
	assert.ElementsMatch(t, expected, actual)
}

func TestIT_MessageRepository_DeleteForRoom(t *testing.T) {
	repo, conn, tx := newTestMessageRepositoryAndTransaction(t)
	defer conn.Close(context.Background())
	room := insertTestRoom(t, conn)
	user1 := insertTestUser(t, conn)
	user2 := insertTestUser(t, conn)
	registerUserInRoom(t, conn, user1.Id, room.Id)
	registerUserInRoom(t, conn, user2.Id, room.Id)

	msg1 := insertTestMessage(t, conn, user1.Id, room.Id)
	msg2 := insertTestMessage(t, conn, user2.Id, room.Id)

	err := repo.DeleteForRoom(context.Background(), tx, room.Id)
	tx.Close(context.Background())
	assert.Nil(t, err, "Actual err: %v", err)

	assertMessageDoesNotExist(t, conn, msg1.Id)
	assertMessageDoesNotExist(t, conn, msg2.Id)
}

func TestIT_MessageRepository_UpdateMessagesOwner(t *testing.T) {
	repo, conn, tx := newTestMessageRepositoryAndTransaction(t)
	defer conn.Close(context.Background())
	userOld := insertTestUser(t, conn)
	userNew := insertTestUser(t, conn)
	room := insertTestRoom(t, conn)
	registerUserInRoom(t, conn, userOld.Id, room.Id)
	registerUserInRoom(t, conn, userNew.Id, room.Id)

	msg := insertTestMessage(t, conn, userOld.Id, room.Id)

	err := repo.UpdateMessagesOwner(context.Background(), tx, userOld.Id, userNew.Name)
	tx.Close(context.Background())
	assert.Nil(t, err, "Actual err: %v", err)

	assertMessageOwner(t, conn, msg.Id, userNew.Id)
	assertUserRegisteredInRoom(t, conn, userOld.Id, room.Id)
}

func TestIT_MessageRepository_UpdateMessagesOwner_DoesNotUpdateOtherUsersMessages(t *testing.T) {
	repo, conn, tx := newTestMessageRepositoryAndTransaction(t)
	defer conn.Close(context.Background())
	userOld := insertTestUser(t, conn)
	userNew := insertTestUser(t, conn)
	user2 := insertTestUser(t, conn)
	room := insertTestRoom(t, conn)
	registerUserInRoom(t, conn, userOld.Id, room.Id)
	registerUserInRoom(t, conn, userNew.Id, room.Id)
	registerUserInRoom(t, conn, user2.Id, room.Id)

	msg := insertTestMessage(t, conn, userOld.Id, room.Id)
	msg2 := insertTestMessage(t, conn, user2.Id, room.Id)

	err := repo.UpdateMessagesOwner(context.Background(), tx, userOld.Id, userNew.Name)
	tx.Close(context.Background())
	assert.Nil(t, err, "Actual err: %v", err)

	assertMessageOwner(t, conn, msg.Id, userNew.Id)
	assertMessageOwner(t, conn, msg2.Id, user2.Id)
}

func TestIT_MessageRepository_UpdateMessagesOwner_UpdatesMessagesInAllRooms(t *testing.T) {
	repo, conn, tx := newTestMessageRepositoryAndTransaction(t)
	defer conn.Close(context.Background())
	userOld := insertTestUser(t, conn)
	userNew := insertTestUser(t, conn)
	room1 := insertTestRoom(t, conn)
	room2 := insertTestRoom(t, conn)
	registerUserInRoom(t, conn, userOld.Id, room1.Id)
	registerUserInRoom(t, conn, userNew.Id, room1.Id)
	registerUserInRoom(t, conn, userOld.Id, room2.Id)
	registerUserInRoom(t, conn, userNew.Id, room2.Id)

	msgRoom1 := insertTestMessage(t, conn, userOld.Id, room1.Id)
	msgRoom2 := insertTestMessage(t, conn, userOld.Id, room2.Id)

	err := repo.UpdateMessagesOwner(context.Background(), tx, userOld.Id, userNew.Name)
	tx.Close(context.Background())
	assert.Nil(t, err, "Actual err: %v", err)

	assertMessageOwner(t, conn, msgRoom1.Id, userNew.Id)
	assertMessageOwner(t, conn, msgRoom2.Id, userNew.Id)
}

func TestIT_MessageRepository_UpdateMessagesOwner_WhenUserDoesNotExist_ExpectNoError(t *testing.T) {
	repo, conn, tx := newTestMessageRepositoryAndTransaction(t)
	defer conn.Close(context.Background())
	userOld := insertTestUser(t, conn)
	userNew := fmt.Sprintf("my-inexistent-user-%s", uuid.NewString())
	room := insertTestRoom(t, conn)
	registerUserInRoom(t, conn, userOld.Id, room.Id)

	msg := insertTestMessage(t, conn, userOld.Id, room.Id)

	err := repo.UpdateMessagesOwner(context.Background(), tx, userOld.Id, userNew)
	tx.Close(context.Background())
	assert.Nil(t, err, "Actual err: %v", err)

	assertMessageOwner(t, conn, msg.Id, userOld.Id)
}

func TestIT_MessageRepository_UpdateMessagesOwner_WhenUserIsNotRegisteredInRoom_ExpectFailure(t *testing.T) {
	repo, conn, tx := newTestMessageRepositoryAndTransaction(t)
	defer conn.Close(context.Background())
	userOld := insertTestUser(t, conn)
	userNew := insertTestUser(t, conn)
	room := insertTestRoom(t, conn)
	registerUserInRoom(t, conn, userOld.Id, room.Id)

	msg := insertTestMessage(t, conn, userOld.Id, room.Id)

	err := repo.UpdateMessagesOwner(context.Background(), tx, userOld.Id, userNew.Name)
	tx.Close(context.Background())
	assert.True(
		t,
		errors.IsErrorWithCode(err, pgx.ForeignKeyValidation),
		"Actual err: %v",
		err,
	)

	assertMessageOwner(t, conn, msg.Id, userOld.Id)
}

func TestIT_MessageRepository_UpdateMessagesOwner_WhenNewUserNotRegisteredAndOldUserHasNoMessage_ExpectSuccess(t *testing.T) {
	repo, conn, tx := newTestMessageRepositoryAndTransaction(t)
	defer conn.Close(context.Background())
	userOld := insertTestUser(t, conn)
	userNew := fmt.Sprintf("my-inexistent-user-%s", uuid.NewString())
	room := insertTestRoom(t, conn)
	registerUserInRoom(t, conn, userOld.Id, room.Id)

	err := repo.UpdateMessagesOwner(context.Background(), tx, userOld.Id, userNew)
	tx.Close(context.Background())
	assert.Nil(t, err, "Actual err: %v", err)
}

func TestIt_MessageRepository_UpdateMessagesOwnerForRoom(t *testing.T) {
	repo, conn, tx := newTestMessageRepositoryAndTransaction(t)
	defer conn.Close(context.Background())
	userOld := insertTestUser(t, conn)
	userNew := insertTestUser(t, conn)
	room := insertTestRoom(t, conn)
	registerUserInRoom(t, conn, userOld.Id, room.Id)
	registerUserInRoom(t, conn, userNew.Id, room.Id)

	msg := insertTestMessage(t, conn, userOld.Id, room.Id)

	err := repo.UpdateMessagesOwnerForRoom(context.Background(), tx, room.Id, userOld.Id, userNew.Name)
	tx.Close(context.Background())
	assert.Nil(t, err, "Actual err: %v", err)

	assertMessageOwner(t, conn, msg.Id, userNew.Id)
	assertUserRegisteredInRoom(t, conn, userOld.Id, room.Id)
}

func TestIT_MessageRepository_UpdateMessagesOwnerForRoom_DoesNotUpdateMessagesFromSameUserInOtherRooms(t *testing.T) {
	repo, conn, tx := newTestMessageRepositoryAndTransaction(t)
	defer conn.Close(context.Background())
	userOld := insertTestUser(t, conn)
	userNew := insertTestUser(t, conn)
	room1 := insertTestRoom(t, conn)
	room2 := insertTestRoom(t, conn)
	registerUserInRoom(t, conn, userOld.Id, room1.Id)
	registerUserInRoom(t, conn, userNew.Id, room1.Id)
	registerUserInRoom(t, conn, userOld.Id, room2.Id)

	msgRoom1 := insertTestMessage(t, conn, userOld.Id, room1.Id)
	msgRoom2 := insertTestMessage(t, conn, userOld.Id, room2.Id)

	err := repo.UpdateMessagesOwnerForRoom(context.Background(), tx, room1.Id, userOld.Id, userNew.Name)
	tx.Close(context.Background())
	assert.Nil(t, err, "Actual err: %v", err)

	assertMessageOwner(t, conn, msgRoom1.Id, userNew.Id)
	assertMessageOwner(t, conn, msgRoom2.Id, userOld.Id)
}

func TestIT_MessageRepository_UpdateMessagesOwnerForRoom_DoesNotUpdateMessagesFromOtherUsersInSameRoom(t *testing.T) {
	repo, conn, tx := newTestMessageRepositoryAndTransaction(t)
	defer conn.Close(context.Background())
	userOld := insertTestUser(t, conn)
	userNew := insertTestUser(t, conn)
	user2 := insertTestUser(t, conn)
	room := insertTestRoom(t, conn)

	registerUserInRoom(t, conn, userOld.Id, room.Id)
	registerUserInRoom(t, conn, userNew.Id, room.Id)
	registerUserInRoom(t, conn, user2.Id, room.Id)

	msg := insertTestMessage(t, conn, user2.Id, room.Id)

	err := repo.UpdateMessagesOwnerForRoom(context.Background(), tx, room.Id, userOld.Id, userNew.Name)
	tx.Close(context.Background())
	assert.Nil(t, err, "Actual err: %v", err)

	assertMessageOwner(t, conn, msg.Id, user2.Id)
}

func TestIT_MessageRepository_UpdateMessagesOwnerForRoom_WhenUserNotRegistered_ExpectSuccess(t *testing.T) {
	repo, conn, tx := newTestMessageRepositoryAndTransaction(t)
	defer conn.Close(context.Background())
	userOld := insertTestUser(t, conn)
	userNew := insertTestUser(t, conn)
	room := insertTestRoom(t, conn)

	err := repo.UpdateMessagesOwnerForRoom(context.Background(), tx, room.Id, userOld.Id, userNew.Name)
	tx.Close(context.Background())
	assert.Nil(t, err, "Actual err: %v", err)
}

func TestIT_MessageRepository_UpdateMessagesOwnerForRoom_WhenNewUserNotRegisteredAndOldUserHasNoMessage_ExpectSuccess(t *testing.T) {
	repo, conn, tx := newTestMessageRepositoryAndTransaction(t)
	defer conn.Close(context.Background())
	userOld := insertTestUser(t, conn)
	userNew := fmt.Sprintf("my-inexistent-user-%s", uuid.NewString())
	room := insertTestRoom(t, conn)
	registerUserInRoom(t, conn, userOld.Id, room.Id)

	err := repo.UpdateMessagesOwnerForRoom(context.Background(), tx, room.Id, userOld.Id, userNew)
	tx.Close(context.Background())
	assert.Nil(t, err, "Actual err: %v", err)
}

func TestIT_MessageRepository_ListForRoom_WhenNoMessageAvailable_ReturnsEmptySlice(t *testing.T) {
	repo, conn := newTestMessageRepository(t)
	defer conn.Close(context.Background())
	room := insertTestRoom(t, conn)

	actual, err := repo.ListForRoom(context.Background(), room.Id)
	assert.Nil(t, err, "Actual err: %v", err)

	assert.Equal(t, []persistence.Message{}, actual)
}

func newTestMessageRepository(t *testing.T) (MessageRepository, db.Connection) {
	conn := newTestConnection(t)
	return NewMessageRepository(conn), conn
}

func newTestMessageRepositoryAndTransaction(t *testing.T) (MessageRepository, db.Connection, db.Transaction) {
	conn := newTestConnection(t)
	tx, err := conn.BeginTx(context.Background())
	assert.Nil(t, err, "Actual err: %v", err)
	return NewMessageRepository(conn), conn, tx
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

func assertMessageOwner(t *testing.T, conn db.Connection, msg uuid.UUID, user uuid.UUID) {
	t.Helper()

	value, err := db.QueryOne[uuid.UUID](
		context.Background(),
		conn,
		"SELECT chat_user FROM message WHERE id = $1",
		msg,
	)
	assert.Nil(t, err, "Actual err: %v", err)
	assert.Equal(t, user, value)
}

func insertTestMessage(
	t *testing.T,
	conn db.Connection,
	user uuid.UUID,
	room uuid.UUID,
) persistence.Message {
	msg := persistence.Message{
		Id:       uuid.New(),
		ChatUser: user,
		Room:     room,
		Message:  "my-message-" + uuid.NewString(),
	}

	createdAt, err := db.QueryOne[time.Time](
		context.Background(),
		conn,
		`INSERT INTO
			message (id, chat_user, room, message)
			VALUES ($1, $2, $3, $4)
			RETURNING created_at`,
		msg.Id,
		msg.ChatUser,
		msg.Room,
		msg.Message,
	)
	assert.Nil(t, err, "Actual err: %v", err)

	msg.CreatedAt = createdAt.UTC()

	return msg
}
