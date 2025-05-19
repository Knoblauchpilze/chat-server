package service

import (
	"context"
	"testing"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/db"
	"github.com/Knoblauchpilze/backend-toolkit/pkg/errors"
	"github.com/Knoblauchpilze/chat-server/pkg/repositories"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestIT_RoomService_RegisterUserInRoom(t *testing.T) {
	service, conn := newTestRegistrationService(t)
	defer conn.Close(context.Background())
	user := insertTestUser(t, conn)
	room := insertTestRoom(t, conn)

	err := service.RegisterUserInRoom(context.Background(), user.Id, room.Id)

	assert.Nil(t, err, "Actual err: %v", err)
	assertUserRegisteredInRoom(t, conn, user.Id, room.Name)
}

func TestIT_RoomService_RegisterUserInRoom_WhenUserDoesNotExist_ExpectError(t *testing.T) {
	service, conn := newTestRegistrationService(t)
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
	service, conn := newTestRegistrationService(t)
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
	service, conn := newTestRegistrationService(t)
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

func newTestRegistrationService(t *testing.T) (RegistrationService, db.Connection) {
	conn := newTestDbConnection(t)
	repos := repositories.New(conn)
	return NewRegistrationService(conn, repos), conn
}
