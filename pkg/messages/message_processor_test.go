package messages

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/db"
	"github.com/Knoblauchpilze/chat-server/pkg/persistence"
	"github.com/Knoblauchpilze/chat-server/pkg/repositories"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestIT_MessageProcessor_EnqueueMessage_ExpectWrittenToDatabase(t *testing.T) {
	processor, dbConn, _ := newTestMessageProcessor(t)
	defer dbConn.Close(context.Background())

	user := insertTestUser(t, dbConn)
	room := insertTestRoom(t, dbConn)
	insertUserInRoom(t, dbConn, user.Id, room.Id)

	wg := asyncStartProcessorAndAssertNoError(t, processor)

	msg := persistence.Message{
		Id:       uuid.New(),
		ChatUser: user.Id,
		Room:     room.Id,
		Message:  fmt.Sprintf("hello %s", room.Name),
	}
	processor.Enqueue(msg)

	err := processor.Stop()
	assert.Nil(t, err, "Actual err: %v", err)
	wg.Wait()

	assertMessageExists(t, dbConn, msg.Id)
}

func TestIT_MessageProcessor_EnqueueMessage_ExpectSentToDispatcher(t *testing.T) {
	processor, dbConn, mock := newTestMessageProcessor(t)
	defer dbConn.Close(context.Background())

	user := insertTestUser(t, dbConn)
	room := insertTestRoom(t, dbConn)
	insertUserInRoom(t, dbConn, user.Id, room.Id)

	wg := asyncStartProcessorAndAssertNoError(t, processor)

	msg := persistence.Message{
		Id:       uuid.New(),
		ChatUser: user.Id,
		Room:     room.Id,
		Message:  fmt.Sprintf("hello %s", room.Name),
	}
	processor.Enqueue(msg)

	err := processor.Stop()
	assert.Nil(t, err, "Actual err: %v", err)
	wg.Wait()

	assert.Equal(t, user.Id, mock.receivedId)
	assert.Equal(t, msg, mock.receivedMsg)
}

func TestIT_MessageProcessor_WhenMessageFailsToBeWritten_ExpectProcessingStops(t *testing.T) {
	testErr := fmt.Errorf("some error")
	mock := newMockMessageRepository(false, testErr)
	repos := repositories.Repositories{
		Message: mock,
	}
	processor := NewMessageProcessor(1, &mockDispatcher{}, repos)

	wg := asyncStartProcessorAndAssertError(t, processor, testErr)

	msg := persistence.Message{}
	processor.Enqueue(msg)

	err := processor.Stop()
	assert.Nil(t, err, "Actual err: %v", err)
	wg.Wait()
}

func newTestMessageProcessor(t *testing.T) (Processor, db.Connection, *mockDispatcher) {
	conn := newTestDbConnection(t)
	mock := &mockDispatcher{}

	repos := repositories.Repositories{
		User:    repositories.NewUserRepository(conn),
		Room:    repositories.NewRoomRepository(conn),
		Message: repositories.NewMessageRepository(conn),
	}

	return NewMessageProcessor(1, mock, repos), conn, mock
}

type mockDispatcher struct {
	receivedId  uuid.UUID
	receivedMsg persistence.Message
}

func (m *mockDispatcher) Broadcast(msg persistence.Message) error {
	m.receivedMsg = msg
	return nil
}

func (m *mockDispatcher) BroadcastExcept(id uuid.UUID, msg persistence.Message) error {
	m.receivedId = id
	m.receivedMsg = msg
	return nil
}

func (m *mockDispatcher) SendTo(id uuid.UUID, msg persistence.Message) {
	m.receivedId = id
	m.receivedMsg = msg
}

type mockMessageRepository struct {
	block   atomic.Bool
	unblock chan struct{}

	err error
}

func newMockMessageRepository(block bool, err error) *mockMessageRepository {
	m := &mockMessageRepository{
		err:     err,
		unblock: make(chan struct{}, 1),
	}

	m.block.Store(block)
	return m
}

func (m *mockMessageRepository) Create(ctx context.Context, msg persistence.Message) (persistence.Message, error) {
	if m.block.Load() {
		<-m.unblock
	}

	return persistence.Message{}, m.err
}

func (m *mockMessageRepository) ListForRoom(ctx context.Context, room uuid.UUID) ([]persistence.Message, error) {
	return []persistence.Message{}, m.err
}
