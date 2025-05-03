package service

import (
	"context"
	"fmt"
	"testing"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/db"
	"github.com/Knoblauchpilze/backend-toolkit/pkg/errors"
	"github.com/Knoblauchpilze/chat-server/pkg/communication"
	"github.com/Knoblauchpilze/chat-server/pkg/persistence"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestIT_MessageService_PostMessage_SendsMessageToProcessor(t *testing.T) {
	service, conn, mock := newTestMessageService(t)
	defer conn.Close(context.Background())
	user := insertTestUser(t, conn)
	room := insertTestRoom(t, conn)
	insertUserInRoom(t, conn, user.Id, room.Id)

	messageDtoRequest := communication.MessageDtoRequest{
		User:    user.Id,
		Room:    room.Id,
		Message: fmt.Sprintf("%s says hello to %s", user.Name, room.Id),
	}

	err := service.PostMessage(context.Background(), messageDtoRequest)

	assert.Nil(t, err, "Actual err: %v", err)
	assert.Len(t, mock.enqueued, 1)
	actual := mock.enqueued[0]
	assert.Equal(t, messageDtoRequest.User, actual.ChatUser)
	assert.Equal(t, messageDtoRequest.Room, actual.Room)
	assert.Equal(t, messageDtoRequest.Message, actual.Message)

}

func TestIT_MessageService_PostMessage_InvalidName(t *testing.T) {
	service, conn, _ := newTestMessageService(t)
	defer conn.Close(context.Background())
	messageDtoRequest := communication.MessageDtoRequest{
		User:    uuid.New(),
		Room:    uuid.New(),
		Message: "",
	}

	err := service.PostMessage(context.Background(), messageDtoRequest)

	assert.True(
		t,
		errors.IsErrorWithCode(err, ErrEmptyMessage),
		"Actual err: %v",
		err,
	)
}

func newTestMessageService(
	t *testing.T,
) (MessageService, db.Connection, *mockProcessor) {
	conn := newTestDbConnection(t)
	mock := &mockProcessor{}
	return NewMessageService(conn, mock), conn, mock
}

type mockProcessor struct {
	enqueued []persistence.Message
}

func (m *mockProcessor) Start() error {
	return nil
}

func (m *mockProcessor) Stop() error {
	return nil
}

func (m *mockProcessor) Enqueue(msg persistence.Message) error {
	m.enqueued = append(m.enqueued, msg)
	return nil
}
