package service

import (
	"context"
	"fmt"
	"testing"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/db"
	"github.com/Knoblauchpilze/backend-toolkit/pkg/errors"
	"github.com/Knoblauchpilze/chat-server/pkg/communication"
	"github.com/Knoblauchpilze/chat-server/pkg/repositories"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestIT_MessageService_PostMessage(t *testing.T) {
	service, conn := newTestMessageService(t)
	defer conn.Close(context.Background())
	user := insertTestUser(t, conn)
	room := insertTestRoom(t, conn)
	insertUserInRoom(t, conn, user.Id, room.Id)

	messageDtoRequest := communication.MessageDtoRequest{
		User:    user.Id,
		Room:    room.Id,
		Message: fmt.Sprintf("%s says hello to %s", user.Name, room.Id),
	}

	out, err := service.PostMessage(context.Background(), messageDtoRequest)

	assert.Nil(t, err, "Actual err: %v", err)

	assert.Equal(t, messageDtoRequest.User, out.User)
	assert.Equal(t, messageDtoRequest.Room, out.Room)
	assert.Equal(t, messageDtoRequest.Message, out.Message)
	assertMessageExists(t, conn, out.Id)
}

func TestIT_MessageService_PostMessage_InvalidName(t *testing.T) {
	service, conn := newTestMessageService(t)
	defer conn.Close(context.Background())
	messageDtoRequest := communication.MessageDtoRequest{
		User:    uuid.New(),
		Room:    uuid.New(),
		Message: "",
	}

	_, err := service.PostMessage(context.Background(), messageDtoRequest)

	assert.True(
		t,
		errors.IsErrorWithCode(err, ErrEmptyMessage),
		"Actual err: %v",
		err,
	)
}

func newTestMessageService(t *testing.T) (MessageService, db.Connection) {
	conn := newTestDbConnection(t)

	repos := repositories.Repositories{
		User:    repositories.NewUserRepository(conn),
		Room:    repositories.NewRoomRepository(conn),
		Message: repositories.NewMessageRepository(conn),
	}

	return NewMessageService(conn, repos), conn
}
