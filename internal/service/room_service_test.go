package service

import (
	"context"
	"fmt"
	"testing"

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
	defer conn.Close(context.Background())
	out, err := service.Create(context.Background(), roomDtoRequest)

	assert.Nil(t, err, "Actual err: %v", err)

	assert.Equal(t, roomDtoRequest.Name, out.Name)
	assertRoomExists(t, conn, out.Id)
}

func TestIT_RoomService_Create_InvalidName(t *testing.T) {
	roomDtoRequest := communication.RoomDtoRequest{
		Name: "",
	}

	service, conn := newTestRoomService(t)
	defer conn.Close(context.Background())
	_, err := service.Create(context.Background(), roomDtoRequest)

	assert.True(
		t,
		errors.IsErrorWithCode(err, ErrInvalidName),
		"Actual err: %v",
		err,
	)
}

func TestIT_RoomService_Create_WhenRoomWithSameNameAlreadyExists_ExpectFailure(t *testing.T) {
	service, conn := newTestRoomService(t)
	defer conn.Close(context.Background())
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
	defer conn.Close(context.Background())
	room := insertTestRoom(t, conn)

	actual, err := service.Get(context.Background(), room.Id)
	assert.Nil(t, err, "Actual err: %v", err)

	expected := communication.ToRoomDtoResponse(room)
	assert.Equal(t, expected, actual)
}

func TestIT_RoomService_Get_WhenRoomDoesNotExist_ExpectFailure(t *testing.T) {
	nonExistingId := uuid.MustParse("00000000-0000-1221-0000-000000000000")

	service, conn := newTestRoomService(t)
	defer conn.Close(context.Background())
	_, err := service.Get(context.Background(), nonExistingId)

	assert.True(
		t,
		errors.IsErrorWithCode(err, db.NoMatchingRows),
		"Actual err: %v",
		err,
	)
}

func TestIT_RoomService_ListUserForRoom_WhenNobodyInRoom_ExpectEmptyList(t *testing.T) {
	service, conn := newTestRoomService(t)
	defer conn.Close(context.Background())

	actual, err := service.ListUserForRoom(context.Background(), uuid.New())

	assert.Nil(t, err, "Actual err: %v", err)
	assert.Equal(t, []communication.UserDtoResponse{}, actual)
}

func TestIT_RoomService_ListUserForRoom(t *testing.T) {
	service, conn := newTestRoomService(t)
	defer conn.Close(context.Background())
	user1 := insertTestUser(t, conn)
	user2 := insertTestUser(t, conn)
	insertTestUser(t, conn)

	room := insertTestRoom(t, conn)

	registerUserInRoom(t, conn, user1.Id, room.Id)
	registerUserInRoom(t, conn, user2.Id, room.Id)

	actual, err := service.ListUserForRoom(context.Background(), room.Id)

	assert.Nil(t, err, "Actual err: %v", err)
	expected := []communication.UserDtoResponse{
		communication.ToUserDtoResponse(user1),
		communication.ToUserDtoResponse(user2),
	}
	assert.ElementsMatch(t, expected, actual)
}

func TestIT_RoomService_ListMessageForRoom_WhenNoMessageInRoom_ExpectEmptyList(t *testing.T) {
	service, conn := newTestRoomService(t)
	defer conn.Close(context.Background())

	actual, err := service.ListMessageForRoom(context.Background(), uuid.New())

	assert.Nil(t, err, "Actual err: %v", err)
	assert.Equal(t, []communication.MessageDtoResponse{}, actual)
}

func TestIT_RoomService_ListMessageForRoom(t *testing.T) {
	service, conn := newTestRoomService(t)
	defer conn.Close(context.Background())
	user1 := insertTestUser(t, conn)
	user2 := insertTestUser(t, conn)
	user3 := insertTestUser(t, conn)

	room1 := insertTestRoom(t, conn)
	room2 := insertTestRoom(t, conn)

	registerUserInRoom(t, conn, user1.Id, room1.Id)
	registerUserInRoom(t, conn, user2.Id, room1.Id)
	registerUserInRoom(t, conn, user3.Id, room2.Id)

	msg1 := insertTestMessage(t, conn, user1.Id, room1.Id)
	msg2 := insertTestMessage(t, conn, user2.Id, room1.Id)
	insertTestMessage(t, conn, user3.Id, room2.Id)

	actual, err := service.ListMessageForRoom(context.Background(), room1.Id)

	assert.Nil(t, err, "Actual err: %v", err)
	expected := []communication.MessageDtoResponse{
		communication.ToMessageDtoResponse(msg1),
		communication.ToMessageDtoResponse(msg2),
	}
	assert.ElementsMatch(t, expected, actual)
}

func TestIT_RoomService_RegisterUserInRoom(t *testing.T) {
	service, conn := newTestRoomService(t)
	defer conn.Close(context.Background())
	user := insertTestUser(t, conn)
	room := insertTestRoom(t, conn)

	err := service.RegisterUserInRoom(context.Background(), user.Id, room.Id)

	assert.Nil(t, err, "Actual err: %v", err)
	assertUserRegisteredInRoom(t, conn, user.Id, room.Name)
}

func TestIT_RoomService_RegisterUserInRoom_WhenUserDoesNotExist_ExpectError(t *testing.T) {
	service, conn := newTestRoomService(t)
	defer conn.Close(context.Background())
	user := uuid.New()
	room := insertTestRoom(t, conn)

	err := service.RegisterUserInRoom(context.Background(), user, room.Id)

	assert.True(
		t,
		errors.IsErrorWithCode(err, repositories.ErrNoSuchUser),
		"Actual err: %v",
		err,
	)
	assertUserNotRegisteredInRoom(t, conn, user, room.Name)
}

func TestIT_RoomService_RegisterUserInRoom_WhenRoomDoesNotExist_ExpectError(t *testing.T) {
	service, conn := newTestRoomService(t)
	defer conn.Close(context.Background())
	user := insertTestUser(t, conn)
	room := uuid.New()

	err := service.RegisterUserInRoom(context.Background(), user.Id, room)

	assert.True(
		t,
		errors.IsErrorWithCode(err, repositories.ErrNoSuchRoom),
		"Actual err: %v",
		err,
	)
}

func TestIT_RoomService_RegisterUserInRoom_WhenUserAlreadyRegistered_ExpectError(t *testing.T) {
	service, conn := newTestRoomService(t)
	defer conn.Close(context.Background())
	user := insertTestUser(t, conn)
	room := insertTestRoom(t, conn)
	registerUserInRoom(t, conn, user.Id, room.Id)

	err := service.RegisterUserInRoom(context.Background(), user.Id, room.Id)

	assert.True(
		t,
		errors.IsErrorWithCode(err, repositories.ErrUserAlreadyRegisteredInRoom),
		"Actual err: %v",
		err,
	)
}

func TestIT_RoomService_Delete(t *testing.T) {
	service, conn := newTestRoomService(t)
	defer conn.Close(context.Background())
	room := insertTestRoom(t, conn)

	err := service.Delete(context.Background(), room.Id)

	assert.Nil(t, err, "Actual err: %v", err)
	assertRoomDoesNotExist(t, conn, room.Id)
}

func TestIT_RoomService_Delete_WhenRoomDoesNotExist_ExpectSuccess(t *testing.T) {
	nonExistingId := uuid.MustParse("00000000-0000-1221-0000-000000000000")

	service, conn := newTestRoomService(t)
	defer conn.Close(context.Background())
	err := service.Delete(context.Background(), nonExistingId)

	assert.Nil(t, err, "Actual err: %v", err)
}

func newTestRoomService(t *testing.T) (RoomService, db.Connection) {
	conn := newTestDbConnection(t)

	repos := repositories.Repositories{
		User:    repositories.NewUserRepository(conn),
		Room:    repositories.NewRoomRepository(conn),
		Message: repositories.NewMessageRepository(conn),
	}

	return NewRoomService(conn, repos), conn
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
