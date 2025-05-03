package messages

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/db"
	"github.com/Knoblauchpilze/chat-server/pkg/communication"
	"github.com/Knoblauchpilze/chat-server/pkg/persistence"
	"github.com/Knoblauchpilze/chat-server/pkg/repositories"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestIT_Processor_EnqueueMessage_ExpectWrittenToDatabase(t *testing.T) {
	processor, dbConn := newTestProcessor(t)
	defer dbConn.Close(context.Background())

	user := insertTestUser(t, dbConn)
	room := insertTestRoom(t, dbConn)
	insertUserInRoom(t, dbConn, user.Id, room.Id)

	wg := asyncStartProcessorAndAssertNoError(t, processor)

	msg := communication.MessageDtoRequest{
		User:    user.Id,
		Room:    room.Id,
		Message: fmt.Sprintf("hello %s", room.Name),
	}
	processor.Enqueue(msg)

	err := processor.Stop()
	assert.Nil(t, err, "Actual err: %v", err)
	wg.Wait()

	assertMessageExists(t, dbConn, msg.User, msg.Room, msg.Message)
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

func TestIT_Processor_WhenMessageQueueIsFull_ExpectCallBlocks(t *testing.T) {
	mock := newMockMessageRepository(true, nil)
	repos := repositories.Repositories{
		Message: mock,
	}
	processor := NewProcessor(1, repos)

	wg := asyncStartProcessorAndAssertNoError(t, processor)

	enqueueMessage := func() {
		msg := communication.MessageDtoRequest{}
		processor.Enqueue(msg)
	}

	// We have a queue of 1: the first message will be enqueued and stuck in
	// the repository. The second one will be enqueued and stay in the queue.
	enqueueMessage()
	enqueueMessage()

	// The third message will not be able to be enqueued
	msgEnqueued := make(chan struct{}, 1)
	go func() {
		defer func() {
			msgEnqueued <- struct{}{}

		}()

		enqueueMessage()
	}()

	timeout := time.After(1 * time.Second)
	select {
	case <-timeout:
	case <-msgEnqueued:
		assert.Fail(t, "Message should not have been enqueued")
	}

	// We need to unblock the message repository
	mock.block.Store(false)
	mock.unblock <- struct{}{}

	err := processor.Stop()
	assert.Nil(t, err, "Actual err: %v", err)
	wg.Wait()
}

func TestIT_Processor_WhenMessageFailsToBeWritten_ExpectProcessingStops(t *testing.T) {
	testErr := fmt.Errorf("some error")
	mock := newMockMessageRepository(false, testErr)
	repos := repositories.Repositories{
		Message: mock,
	}
	processor := NewProcessor(1, repos)

	wg := asyncStartProcessorAndAssertError(t, processor, testErr)

	msg := communication.MessageDtoRequest{}
	processor.Enqueue(msg)

	err := processor.Stop()
	assert.Nil(t, err, "Actual err: %v", err)
	wg.Wait()
}

func newTestProcessor(t *testing.T) (Processor, db.Connection) {
	conn := newTestDbConnection(t)

	repos := repositories.Repositories{
		User:    repositories.NewUserRepository(conn),
		Room:    repositories.NewRoomRepository(conn),
		Message: repositories.NewMessageRepository(conn),
	}

	return NewProcessor(1, repos), conn
}

func asyncStartProcessorAndAssertNoError(
	t *testing.T, processor Processor,
) *sync.WaitGroup {
	return asyncStartProcessorAndAssertError(t, processor, nil)
}

func asyncStartProcessorAndAssertError(
	t *testing.T, processor Processor, expectedErr error,
) *sync.WaitGroup {
	var wg sync.WaitGroup

	wg.Add(1)

	go func() {
		defer wg.Done()
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Processor panicked: %v", r)
			}
		}()

		err := processor.Start()
		assert.Equal(t, expectedErr, err, "Actual err: %v", err)
	}()

	// Wait a bit for the processor to start
	time.Sleep(50 * time.Millisecond)

	return &wg
}
