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

func TestIT_RegistrationService_RegisterUserInRoom(t *testing.T) {
	service, conn := newTestRegistrationService(t)
	defer conn.Close(context.Background())
	user := insertTestUser(t, conn)
	room := insertTestRoom(t, conn)

	err := service.RegisterUserInRoom(context.Background(), user.Id, room.Id)

	assert.Nil(t, err, "Actual err: %v", err)
	assertUserRegisteredInRoom(t, conn, user.Id, room.Name)
}

func TestIT_RegistrationService_RegisterUserInRoom_WhenUserDoesNotExist_ExpectError(t *testing.T) {
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

func TestIT_RegistrationService_RegisterUserInRoom_WhenRoomDoesNotExist_ExpectError(t *testing.T) {
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

func TestIT_RegistrationService_RegisterUserInRoom_WhenUserAlreadyRegistered_ExpectError(t *testing.T) {
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

func TestIT_RegistrationService_UnregisterUserInRoom(t *testing.T) {
	service, conn := newTestRegistrationService(t)
	defer conn.Close(context.Background())
	user1 := insertTestUser(t, conn)
	user2 := insertTestUser(t, conn)
	room1 := insertTestRoom(t, conn)
	room2 := insertTestRoom(t, conn)
	registerUserInRoom(t, conn, user1.Id, room1.Id)
	registerUserInRoom(t, conn, user1.Id, room2.Id)
	registerUserInRoom(t, conn, user2.Id, room1.Id)

	err := service.UnregisterUserInRoom(context.Background(), user1.Id, room1.Id)

	assert.Nil(t, err, "Actual err: %v", err)
	assertUserNotRegisteredInRoom(t, conn, user1.Id, room1.Name)
	assertUserRegisteredInRoom(t, conn, user1.Id, room2.Name)
	assertUserRegisteredInRoom(t, conn, user2.Id, room1.Name)
}

func TestIT_RegistrationService_ShouldNotUnregisterFromGeneralRoom(t *testing.T) {
	service, conn := newTestRegistrationService(t)
	defer conn.Close(context.Background())
	user := insertTestUser(t, conn)
	room := getRoomId(t, conn, "general")

	err := service.UnregisterUserInRoom(context.Background(), user.Id, room)

	assert.True(
		t,
		errors.IsErrorWithCode(err, ErrLeavingRoomIsNotAllowed),
		"Actual err: %v",
		err,
	)
}

func TestIT_RegistrationService_UnregisterUserInRoom_UpdateMessagesInRoom(t *testing.T) {
	service, conn := newTestRegistrationService(t)
	defer conn.Close(context.Background())
	user1 := insertTestUser(t, conn)
	user2 := insertTestUser(t, conn)
	room := insertTestRoom(t, conn)
	registerUserInRoom(t, conn, user1.Id, room.Id)
	registerUserInRoom(t, conn, user2.Id, room.Id)
	registerUserByNameInRoom(t, conn, "ghost", room.Id)
	msg1 := insertTestMessage(t, conn, user1.Id, room.Id)
	msg2 := insertTestMessage(t, conn, user2.Id, room.Id)

	err := service.UnregisterUserInRoom(context.Background(), user1.Id, room.Id)

	assert.Nil(t, err, "Actual err: %v", err)
	assertMessageOwner(t, conn, msg1.Id, "ghost")
	assertMessageOwner(t, conn, msg2.Id, user2.Name)
}

func newTestRegistrationService(t *testing.T) (RegistrationService, db.Connection) {
	conn := newTestDbConnection(t)
	repos := repositories.New(conn)
	return NewRegistrationService(conn, repos), conn
}
