package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/db"
	"github.com/Knoblauchpilze/backend-toolkit/pkg/db/pgx"
	"github.com/Knoblauchpilze/backend-toolkit/pkg/errors"
	"github.com/Knoblauchpilze/chat-server/pkg/communication"
	"github.com/Knoblauchpilze/chat-server/pkg/persistence"
	"github.com/Knoblauchpilze/chat-server/pkg/repositories"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestIT_RoomService_Create(t *testing.T) {
	id := uuid.New()
	roomDtoRequest := communication.RoomDtoRequest{
		Name: fmt.Sprintf("my-room-%s", id),
	}

	service, conn := newTestRoomService(t)
	out, err := service.Create(context.Background(), roomDtoRequest)

	assert.Nil(t, err, "Actual err: %v", err)

	assert.Equal(t, roomDtoRequest.Name, out.Name)
	assertRoomExists(t, conn, out.Id)
}

func TestIT_RoomService_Create_InvalidName(t *testing.T) {
	roomDtoRequest := communication.RoomDtoRequest{
		Name: "",
	}

	service, _ := newTestRoomService(t)
	_, err := service.Create(context.Background(), roomDtoRequest)

	assert.True(
		t,
		errors.IsErrorWithCode(err, InvalidName),
		"Actual err: %v",
		err,
	)
}

func TestIT_RoomService_Create_WhenRoomWithSameNameAlreadyExists_ExpectFailure(t *testing.T) {
	service, conn := newTestRoomService(t)
	room := insertTestRoom(t, conn)
	roomDtoRequest := communication.RoomDtoRequest{
		Name: room.Name,
	}

	_, err := service.Create(context.Background(), roomDtoRequest)

	assert.True(
		t,
		errors.IsErrorWithCode(err, pgx.UniqueConstraintViolation),
		"Actual err: %v",
		err,
	)
}

func TestIT_RoomService_Get(t *testing.T) {
	service, conn := newTestRoomService(t)
	room := insertTestRoom(t, conn)

	actual, err := service.Get(context.Background(), room.Id)

	assert.Nil(t, err, "Actual err: %v", err)
	assert.Equal(t, room.Id, actual.Id)
	assert.Equal(t, room.Name, actual.Name)
}

func TestIT_RoomService_Get_WhenRoomDoesNotExist_ExpectFailure(t *testing.T) {
	nonExistingId := uuid.MustParse("00000000-0000-1221-0000-000000000000")

	service, _ := newTestRoomService(t)
	_, err := service.Get(context.Background(), nonExistingId)

	assert.True(
		t,
		errors.IsErrorWithCode(err, db.NoMatchingRows),
		"Actual err: %v",
		err,
	)
}

func TestIT_RoomService_Delete(t *testing.T) {
	service, conn := newTestRoomService(t)
	room := insertTestRoom(t, conn)

	err := service.Delete(context.Background(), room.Id)

	assert.Nil(t, err, "Actual err: %v", err)
	assertRoomDoesNotExist(t, conn, room.Id)
}

func TestIT_RoomService_Delete_WhenRoomDoesNotExist_ExpectSuccess(t *testing.T) {
	nonExistingId := uuid.MustParse("00000000-0000-1221-0000-000000000000")

	service, _ := newTestRoomService(t)
	err := service.Delete(context.Background(), nonExistingId)

	assert.Nil(t, err, "Actual err: %v", err)
}

func newTestRoomService(t *testing.T) (RoomService, db.Connection) {
	conn := newTestDbConnection(t)

	repos := repositories.Repositories{
		Room: repositories.NewRoomRepository(conn),
	}

	return NewRoomService(conn, repos), conn
}

func insertTestRoom(t *testing.T, conn db.Connection) persistence.Room {
	repo := repositories.NewRoomRepository(conn)

	id := uuid.New()
	room := persistence.Room{
		Id:        id,
		Name:      fmt.Sprintf("my-room-%s", id),
		CreatedAt: time.Now(),
	}
	out, err := repo.Create(context.Background(), room)
	assert.Nil(t, err, "Actual err: %v", err)

	assertRoomExists(t, conn, out.Id)

	return out
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
