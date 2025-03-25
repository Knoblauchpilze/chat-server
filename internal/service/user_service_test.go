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
	eassert "github.com/Knoblauchpilze/easy-assert/assert"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestIT_UserService_Create(t *testing.T) {
	id := uuid.New()
	userDtoRequest := communication.UserDtoRequest{
		Name: fmt.Sprintf("my-user-%s", id),
	}

	service, conn := newTestUserService(t)
	out, err := service.Create(context.Background(), userDtoRequest)

	assert.Nil(t, err, "Actual err: %v", err)

	assert.Equal(t, userDtoRequest.Name, out.Name)
	assertUserExists(t, conn, out.Id)
}

func TestIT_UserService_Create_InvalidName(t *testing.T) {
	userDtoRequest := communication.UserDtoRequest{
		Name: "",
	}

	service, _ := newTestUserService(t)
	_, err := service.Create(context.Background(), userDtoRequest)

	assert.True(
		t,
		errors.IsErrorWithCode(err, InvalidName),
		"Actual err: %v",
		err,
	)
}

func TestIT_UserService_Create_WhenUserWithSameNameAlreadyExists_ExpectFailure(t *testing.T) {
	service, conn := newTestUserService(t)
	user := insertTestUser(t, conn)
	userDtoRequest := communication.UserDtoRequest{
		Name:    user.Name,
		ApiUser: uuid.New(),
	}

	_, err := service.Create(context.Background(), userDtoRequest)

	assert.True(
		t,
		errors.IsErrorWithCode(err, pgx.UniqueConstraintViolation),
		"Actual err: %v",
		err,
	)
}

func TestIT_UserService_Get(t *testing.T) {
	service, conn := newTestUserService(t)
	user := insertTestUser(t, conn)

	actual, err := service.Get(context.Background(), user.Id)

	assert.Nil(t, err, "Actual err: %v", err)
	assert.Equal(t, user.Id, actual.Id)
	assert.Equal(t, user.Name, actual.Name)
	assert.Equal(t, user.ApiUser, actual.ApiUser)
}

func TestIT_UserService_Get_WhenUserDoesNotExist_ExpectFailure(t *testing.T) {
	nonExistingId := uuid.MustParse("00000000-0000-1221-0000-000000000000")

	service, _ := newTestUserService(t)
	_, err := service.Get(context.Background(), nonExistingId)

	assert.True(
		t,
		errors.IsErrorWithCode(err, db.NoMatchingRows),
		"Actual err: %v",
		err,
	)
}

func TestIT_UserService_ListForUser_WhenNoRoomRegistered_ExpectEmptyList(t *testing.T) {
	service, _ := newTestUserService(t)

	actual, err := service.ListForUser(context.Background(), uuid.New())

	assert.Nil(t, err, "Actual err: %v", err)
	assert.Equal(t, []communication.RoomDtoResponse{}, actual)
}

func TestIT_UserService_ListForUser(t *testing.T) {
	service, conn := newTestUserService(t)
	user := insertTestUser(t, conn)

	room1 := insertTestRoom(t, conn)
	insertTestRoom(t, conn)

	insertUserInRoom(t, conn, user.Id, room1.Id)

	actual, err := service.ListForUser(context.Background(), user.Id)
	assert.Nil(t, err, "Actual err: %v", err)

	assert.Len(t, actual, 1)
	expected := communication.ToRoomDtoResponse(room1)
	assert.True(t, eassert.EqualsIgnoringFields(actual[0], expected))
}

func TestIT_UserService_Delete(t *testing.T) {
	service, conn := newTestUserService(t)
	user := insertTestUser(t, conn)

	err := service.Delete(context.Background(), user.Id)

	assert.Nil(t, err, "Actual err: %v", err)
	assertUserDoesNotExist(t, conn, user.Id)
}

func TestIT_UserService_Delete_WhenUserDoesNotExist_ExpectSuccess(t *testing.T) {
	nonExistingId := uuid.MustParse("00000000-0000-1221-0000-000000000000")

	service, _ := newTestUserService(t)
	err := service.Delete(context.Background(), nonExistingId)

	assert.Nil(t, err, "Actual err: %v", err)
}

func newTestUserService(t *testing.T) (UserService, db.Connection) {
	conn := newTestDbConnection(t)

	repos := repositories.Repositories{
		Room: repositories.NewRoomRepository(conn),
		User: repositories.NewUserRepository(conn),
	}

	return NewUserService(conn, repos), conn
}

func insertTestUser(t *testing.T, conn db.Connection) persistence.User {
	repo := repositories.NewUserRepository(conn)

	id := uuid.New()
	user := persistence.User{
		Id:        id,
		Name:      fmt.Sprintf("my-user-%s", id),
		ApiUser:   uuid.New(),
		CreatedAt: time.Now(),
	}
	out, err := repo.Create(context.Background(), user)
	assert.Nil(t, err, "Actual err: %v", err)

	assertUserExists(t, conn, out.Id)

	return out
}

func assertUserExists(t *testing.T, conn db.Connection, id uuid.UUID) {
	value, err := db.QueryOne[uuid.UUID](
		context.Background(),
		conn,
		"SELECT id FROM chat_user WHERE id = $1",
		id,
	)
	assert.Nil(t, err, "Actual err: %v", err)
	assert.Equal(t, id, value)
}

func assertUserDoesNotExist(t *testing.T, conn db.Connection, id uuid.UUID) {
	value, err := db.QueryOne[int](
		context.Background(),
		conn,
		"SELECT COUNT(id) FROM chat_user WHERE id = $1",
		id,
	)
	assert.Nil(t, err, "Actual err: %v", err)
	assert.Zero(t, value)
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
