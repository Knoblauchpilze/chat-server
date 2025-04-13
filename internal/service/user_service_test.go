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

func TestIT_UserService_Create(t *testing.T) {
	id := uuid.New()
	userDtoRequest := communication.UserDtoRequest{
		Name: fmt.Sprintf("my-user-%s", id),
	}

	service, conn := newTestUserService(t)
	defer conn.Close(context.Background())
	out, err := service.Create(context.Background(), userDtoRequest)

	assert.Nil(t, err, "Actual err: %v", err)

	assert.Equal(t, userDtoRequest.Name, out.Name)
	assertUserExists(t, conn, out.Id)
}

func TestIT_UserService_Create_InvalidName(t *testing.T) {
	userDtoRequest := communication.UserDtoRequest{
		Name: "",
	}

	service, conn := newTestUserService(t)
	defer conn.Close(context.Background())
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
	defer conn.Close(context.Background())
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

func TestIT_UserService_Create_RegistersUserInGeneralRoomt(t *testing.T) {
	id := uuid.New()
	userDtoRequest := communication.UserDtoRequest{
		Name: fmt.Sprintf("my-user-%s", id),
	}

	service, conn := newTestUserService(t)
	defer conn.Close(context.Background())
	out, err := service.Create(context.Background(), userDtoRequest)

	assert.Nil(t, err, "Actual err: %v", err)

	assertUserRegisteredInRoom(t, conn, out.Id, "general")
}

func TestIT_UserService_Get(t *testing.T) {
	service, conn := newTestUserService(t)
	defer conn.Close(context.Background())
	user := insertTestUser(t, conn)

	actual, err := service.Get(context.Background(), user.Id)
	assert.Nil(t, err, "Actual err: %v", err)

	expected := communication.ToUserDtoResponse(user)
	assert.Equal(t, expected, actual)
}

func TestIT_UserService_Get_WhenUserDoesNotExist_ExpectFailure(t *testing.T) {
	nonExistingId := uuid.MustParse("00000000-0000-1221-0000-000000000000")

	service, conn := newTestUserService(t)
	defer conn.Close(context.Background())
	_, err := service.Get(context.Background(), nonExistingId)

	assert.True(
		t,
		errors.IsErrorWithCode(err, db.NoMatchingRows),
		"Actual err: %v",
		err,
	)
}

func TestIT_UserService_GetByName(t *testing.T) {
	service, conn := newTestUserService(t)
	defer conn.Close(context.Background())
	user1 := insertTestUser(t, conn)
	insertTestUser(t, conn)

	actual, err := service.GetByName(context.Background(), user1.Name)
	assert.Nil(t, err, "Actual err: %v", err)

	expected := communication.ToUserDtoResponse(user1)
	assert.Equal(t, expected, actual)
}

func TestIT_UserService_GetByName_WhenUserDoesNotExist_ExpectFailure(t *testing.T) {
	nonExistingName := "my-non-existent-name"

	service, conn := newTestUserService(t)
	defer conn.Close(context.Background())
	_, err := service.GetByName(context.Background(), nonExistingName)

	assert.True(
		t,
		errors.IsErrorWithCode(err, db.NoMatchingRows),
		"Actual err: %v",
		err,
	)
}

func TestIT_UserService_ListForUser_WhenNoRoomRegistered_ExpectEmptyList(t *testing.T) {
	service, conn := newTestUserService(t)
	defer conn.Close(context.Background())

	actual, err := service.ListForUser(context.Background(), uuid.New())

	assert.Nil(t, err, "Actual err: %v", err)
	assert.Equal(t, []communication.RoomDtoResponse{}, actual)
}

func TestIT_UserService_ListForUser(t *testing.T) {
	service, conn := newTestUserService(t)
	defer conn.Close(context.Background())
	user := insertTestUser(t, conn)

	room1 := insertTestRoom(t, conn)
	insertTestRoom(t, conn)

	insertUserInRoom(t, conn, user.Id, room1.Id)

	actual, err := service.ListForUser(context.Background(), user.Id)
	assert.Nil(t, err, "Actual err: %v", err)

	expected := []communication.RoomDtoResponse{
		communication.ToRoomDtoResponse(room1),
	}
	assert.Equal(t, expected, actual)
}

func TestIT_UserService_Delete(t *testing.T) {
	service, conn := newTestUserService(t)
	defer conn.Close(context.Background())
	user := insertTestUser(t, conn)

	err := service.Delete(context.Background(), user.Id)

	assert.Nil(t, err, "Actual err: %v", err)
	assertUserDoesNotExist(t, conn, user.Id)
}

func TestIT_UserService_Delete_WhenUserDoesNotExist_ExpectSuccess(t *testing.T) {
	nonExistingId := uuid.MustParse("00000000-0000-1221-0000-000000000000")

	service, conn := newTestUserService(t)
	defer conn.Close(context.Background())
	err := service.Delete(context.Background(), nonExistingId)

	assert.Nil(t, err, "Actual err: %v", err)
}

func TestIT_UserService_Delete_RemovesUserFromGeneralRoomt(t *testing.T) {
	service, conn := newTestUserService(t)
	defer conn.Close(context.Background())
	user := insertTestUser(t, conn)

	err := service.Delete(context.Background(), user.Id)

	assert.Nil(t, err, "Actual err: %v", err)
	assertUserNotRegisteredInRoom(t, conn, user.Id, "general")
}

func newTestUserService(t *testing.T) (UserService, db.Connection) {
	conn := newTestDbConnection(t)

	repos := repositories.Repositories{
		User: repositories.NewUserRepository(conn),
		Room: repositories.NewRoomRepository(conn),
	}

	return NewUserService(conn, repos), conn
}

func insertTestUser(t *testing.T, conn db.Connection) persistence.User {
	repo := repositories.NewUserRepository(conn)

	tx, err := conn.BeginTx(context.Background())
	assert.Nil(t, err, "Actual err: %v", err)

	id := uuid.New()
	user := persistence.User{
		Id:      id,
		Name:    fmt.Sprintf("my-user-%s", id),
		ApiUser: uuid.New(),
	}
	out, err := repo.Create(context.Background(), tx, user)
	tx.Close(context.Background())
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

func assertUserNotRegisteredInRoom(
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
	assert.Equal(t, 0, value)
}
