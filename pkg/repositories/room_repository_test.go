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

func TestIT_RoomRepository_Create(t *testing.T) {
	beforeInsertion := time.Now()

	repo, conn, tx := newTestRoomRepositoryAndTransaction(t)
	defer conn.Close(context.Background())

	room := persistence.Room{
		Id:   uuid.New(),
		Name: "my-room-" + uuid.New().String(),
	}

	actual, err := repo.Create(context.Background(), tx, room)
	tx.Close(context.Background())
	assert.Nil(t, err, "Actual err: %v", err)

	assert.True(t, eassert.EqualsIgnoringFields(actual, room, "CreatedAt", "UpdatedAt"))
	assert.Equal(t, actual.CreatedAt, actual.UpdatedAt)
	assert.True(t, actual.CreatedAt.After(beforeInsertion))
	assertRoomExists(t, conn, room.Id)
}

func TestIT_RoomRepository_Create_WhenDuplicateName_ExpectFailure(t *testing.T) {
	repo, conn, tx := newTestRoomRepositoryAndTransaction(t)
	defer conn.Close(context.Background())
	room := insertTestRoom(t, conn)

	newRoom := persistence.Room{
		Id:   uuid.New(),
		Name: room.Name,
	}

	_, err := repo.Create(context.Background(), tx, newRoom)
	tx.Close(context.Background())

	assert.True(
		t,
		errors.IsErrorWithCode(err, pgx.UniqueConstraintViolation),
		"Actual err: %v",
		err,
	)
	assertRoomDoesNotExist(t, conn, newRoom.Id)
}

func TestIT_RoomRepository_Get(t *testing.T) {
	repo, conn := newTestRoomRepository(t)
	defer conn.Close(context.Background())
	room := insertTestRoom(t, conn)

	actual, err := repo.Get(context.Background(), room.Id)
	assert.Nil(t, err, "Actual err: %v", err)

	assert.Equal(t, room, actual)
}

func TestIT_RoomRepository_Get_WhenNotFound_ExpectFailure(t *testing.T) {
	repo, conn := newTestRoomRepository(t)
	defer conn.Close(context.Background())

	// Non-existent id
	id := uuid.MustParse("00000000-1111-2222-1111-000000000000")
	_, err := repo.Get(context.Background(), id)
	assert.True(
		t,
		errors.IsErrorWithCode(err, db.NoMatchingRows),
		"Actual err: %v",
		err,
	)
}

func TestIT_RoomRepository_List(t *testing.T) {
	repo, conn := newTestRoomRepository(t)
	defer conn.Close(context.Background())
	room := insertTestRoom(t, conn)

	rooms, err := repo.List(context.Background())
	assert.Nil(t, err, "Actual err: %v", err)

	assert.Contains(t, rooms, room)
}

func TestIT_RoomRepository_UserInRoom(t *testing.T) {
	repo, conn := newTestRoomRepository(t)
	defer conn.Close(context.Background())
	user := insertTestUser(t, conn)
	room := insertTestRoom(t, conn)
	registerUserInRoom(t, conn, user.Id, room.Id)

	actual, err := repo.UserInRoom(context.Background(), user.Id, room.Id)
	assert.Nil(t, err, "Actual err: %v", err)
	assert.True(t, actual)
}

func TestIT_RoomRepository_UserInRoom_WhenNotRegistered_ExpectFalse(t *testing.T) {
	repo, conn := newTestRoomRepository(t)
	defer conn.Close(context.Background())
	user := insertTestUser(t, conn)
	room := insertTestRoom(t, conn)

	actual, err := repo.UserInRoom(context.Background(), user.Id, room.Id)
	assert.Nil(t, err, "Actual err: %v", err)
	assert.False(t, actual)
}

func TestIT_RoomRepository_ListForUser(t *testing.T) {
	repo, conn := newTestRoomRepository(t)
	defer conn.Close(context.Background())
	user := insertTestUser(t, conn)
	room1 := insertTestRoom(t, conn)
	insertTestRoom(t, conn)
	registerUserInRoom(t, conn, user.Id, room1.Id)

	actual, err := repo.ListForUser(context.Background(), user.Id)
	assert.Nil(t, err, "Actual err: %v", err)

	expected := []persistence.Room{room1}
	assert.ElementsMatch(t, expected, actual)
}

func TestIT_RoomRepository_ListForUser_WhenNoRoomRegistered_ReturnsEmptySlice(t *testing.T) {
	repo, conn := newTestRoomRepository(t)
	defer conn.Close(context.Background())
	user := insertTestUser(t, conn)

	actual, err := repo.ListForUser(context.Background(), user.Id)
	assert.Nil(t, err, "Actual err: %v", err)

	assert.Equal(t, []persistence.Room{}, actual)
}

func TestIT_RoomRepository_Delete(t *testing.T) {
	repo, conn, tx := newTestRoomRepositoryAndTransaction(t)
	defer conn.Close(context.Background())

	room := insertTestRoom(t, conn)

	err := repo.Delete(context.Background(), tx, room.Id)
	tx.Close(context.Background())

	assert.Nil(t, err, "Actual err: %v", err)
	assertRoomDoesNotExist(t, conn, room.Id)
}

func TestIT_RoomRepository_Delete_WhenNotFound_ExpectSuccess(t *testing.T) {
	repo, conn, tx := newTestRoomRepositoryAndTransaction(t)
	defer conn.Close(context.Background())

	room := insertTestRoom(t, conn)
	id := uuid.New()
	assert.NotEqual(t, room.Id, id)

	err := repo.Delete(context.Background(), tx, id)
	tx.Close(context.Background())

	assert.Nil(t, err, "Actual err: %v", err)
	assertRoomExists(t, conn, room.Id)
}

func newTestRoomRepository(t *testing.T) (RoomRepository, db.Connection) {
	conn := newTestConnection(t)
	return NewRoomRepository(conn), conn
}

func newTestRoomRepositoryAndTransaction(t *testing.T) (RoomRepository, db.Connection, db.Transaction) {
	conn := newTestConnection(t)
	tx, err := conn.BeginTx(context.Background())
	assert.Nil(t, err, "Actual err: %v", err)
	return NewRoomRepository(conn), conn, tx
}

func assertRoomExists(t *testing.T, conn db.Connection, id uuid.UUID) {
	value, err := db.QueryOne[uuid.UUID](
		context.Background(),
		conn,
		"SELECT id FROM room WHERE id = $1",
		id,
	)
	assert.Nil(t, err, "Actual err: %v", err)
	assert.Equal(t, id, value)
}

func assertRoomDoesNotExist(t *testing.T, conn db.Connection, id uuid.UUID) {
	value, err := db.QueryOne[int](
		context.Background(),
		conn,
		"SELECT COUNT(id) FROM room WHERE id = $1",
		id,
	)
	assert.Nil(t, err, "Actual err: %v", err)
	assert.Zero(t, value)
}

func insertTestRoom(t *testing.T, conn db.Connection) persistence.Room {
	room := persistence.Room{
		Id:   uuid.New(),
		Name: "my-room-" + uuid.New().String(),
	}

	times, err := db.QueryOne[createdAtUpdatedAt](
		context.Background(),
		conn,
		`INSERT INTO
			room (id, name)
			VALUES ($1, $2)
			RETURNING created_at, updated_at`,
		room.Id,
		room.Name,
	)
	assert.Nil(t, err, "Actual err: %v", err)

	room.CreatedAt = times.CreatedAt.UTC()
	room.UpdatedAt = times.UpdatedAt.UTC()

	return room
}
