package clients

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/db"
	"github.com/Knoblauchpilze/backend-toolkit/pkg/errors"
	"github.com/Knoblauchpilze/chat-server/pkg/persistence"
	"github.com/Knoblauchpilze/chat-server/pkg/repositories"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestUnit_Manager_WhenClientAlreadyRegistered_ExpectError(t *testing.T) {
	manager, dbConn := newTestManager(t)
	defer dbConn.Close(context.Background())
	id := uuid.New()

	err := manager.OnConnect(id, nil)
	assert.Nil(t, err, "Actual err: %v", err)

	err = manager.OnConnect(id, nil)
	assert.True(
		t,
		errors.IsErrorWithCode(err, ErrClientAlreadyRegistered),
		"Actual err: %v",
		err,
	)
}

func TestUnit_Manager_WhenClosing_ExpectClientIsAlsoClosed(t *testing.T) {
	manager, dbConn := newTestManager(t)
	defer dbConn.Close(context.Background())
	id := uuid.New()
	mock := &mockClient{}

	err := manager.OnConnect(id, mock)
	assert.Nil(t, err, "Actual err: %v", err)

	asyncStartManagerAndAssertNoError(t, manager)

	err = manager.Stop()
	assert.Nil(t, err, "Actual err: %v", err)

	assert.Equal(t, 1, mock.stopCalled)
}

func TestUnit_Manager_WhenUserInRoomAndBroadcast_ExpectMessageReceived(t *testing.T) {
	manager, dbConn := newTestManager(t)
	defer dbConn.Close(context.Background())
	mock := &mockClient{}

	user1 := insertTestUser(t, dbConn)
	room := insertTestRoom(t, dbConn)
	registerUserInRoom(t, dbConn, user1.Id, room.Id)

	err := manager.OnConnect(user1.Id, mock)
	assert.Nil(t, err, "Actual err: %v", err)

	msg := persistence.Message{
		Id:        uuid.New(),
		ChatUser:  uuid.New(),
		Room:      room.Id,
		Message:   "Hello",
		CreatedAt: time.Date(2025, 5, 5, 21, 44, 20, 0, time.UTC),
	}
	manager.Broadcast(msg)

	assert.Equal(t, 1, mock.enqueueCalled)
	expected := []persistence.Message{msg}
	assert.Equal(t, expected, mock.enqueued, 1)
}

func TestUnit_Manager_WhenUserNotInRoomAndBroadcast_ExpectMessageNotReceived(t *testing.T) {
	manager, dbConn := newTestManager(t)
	defer dbConn.Close(context.Background())
	mock := &mockClient{}

	user1 := insertTestUser(t, dbConn)
	room := insertTestRoom(t, dbConn)

	err := manager.OnConnect(user1.Id, mock)
	assert.Nil(t, err, "Actual err: %v", err)

	msg := persistence.Message{
		Id:        uuid.New(),
		ChatUser:  uuid.New(),
		Room:      room.Id,
		Message:   "Hello",
		CreatedAt: time.Date(2025, 5, 5, 21, 44, 20, 0, time.UTC),
	}
	manager.Broadcast(msg)

	assert.Equal(t, 0, mock.enqueueCalled)
}

func TestUnit_Manager_WhenBroadcastAfterDisconnect_ExpectNoMessageReceived(t *testing.T) {
	manager, dbConn := newTestManager(t)
	defer dbConn.Close(context.Background())
	mock := &mockClient{}

	user1 := insertTestUser(t, dbConn)
	room := insertTestRoom(t, dbConn)
	registerUserInRoom(t, dbConn, user1.Id, room.Id)

	err := manager.OnConnect(user1.Id, mock)
	assert.Nil(t, err, "Actual err: %v", err)
	manager.OnDisconnect(user1.Id)

	msg := persistence.Message{
		Id:        uuid.New(),
		ChatUser:  uuid.New(),
		Room:      room.Id,
		Message:   "Hello",
		CreatedAt: time.Date(2025, 5, 5, 21, 44, 20, 0, time.UTC),
	}
	manager.Broadcast(msg)

	assert.Equal(t, 0, mock.enqueueCalled)
}

func TestUnit_Manager_WhenBroadcastExceptToClient_ExpectNoMessageReceived(t *testing.T) {
	manager, dbConn := newTestManager(t)
	defer dbConn.Close(context.Background())
	mock1 := &mockClient{}
	mock2 := &mockClient{}

	user1 := insertTestUser(t, dbConn)
	user2 := insertTestUser(t, dbConn)
	room := insertTestRoom(t, dbConn)
	registerUserInRoom(t, dbConn, user1.Id, room.Id)
	registerUserInRoom(t, dbConn, user2.Id, room.Id)

	err := manager.OnConnect(user1.Id, mock1)
	assert.Nil(t, err, "Actual err: %v", err)
	err = manager.OnConnect(user2.Id, mock2)
	assert.Nil(t, err, "Actual err: %v", err)

	msg := persistence.Message{
		Id:        uuid.New(),
		ChatUser:  uuid.New(),
		Room:      room.Id,
		Message:   "Hello",
		CreatedAt: time.Date(2025, 5, 5, 21, 44, 20, 0, time.UTC),
	}
	manager.BroadcastExcept(user2.Id, msg)

	assert.Equal(t, 1, mock1.enqueueCalled)
	expected := []persistence.Message{msg}
	assert.Equal(t, expected, mock1.enqueued, 1)

	assert.Equal(t, 0, mock2.enqueueCalled)
}

func TestUnit_Manager_WhenSendingMessageToSpecificClient_ExpectMessageReceived(t *testing.T) {
	manager, dbConn := newTestManager(t)
	defer dbConn.Close(context.Background())
	clientId1 := uuid.New()
	mock1 := &mockClient{}
	clientId2 := uuid.New()
	mock2 := &mockClient{}

	err := manager.OnConnect(clientId1, mock1)
	assert.Nil(t, err, "Actual err: %v", err)
	err = manager.OnConnect(clientId2, mock2)
	assert.Nil(t, err, "Actual err: %v", err)

	msg := persistence.Message{
		Id:        uuid.New(),
		ChatUser:  uuid.New(),
		Room:      uuid.New(),
		Message:   "Hello",
		CreatedAt: time.Date(2025, 5, 5, 21, 44, 20, 0, time.UTC),
	}
	manager.SendTo(clientId1, msg)

	assert.Equal(t, 1, mock1.enqueueCalled)
	expected := []persistence.Message{msg}
	assert.Equal(t, expected, mock1.enqueued, 1)

	assert.Equal(t, 0, mock2.enqueueCalled)
}

func newTestManager(t *testing.T) (Manager, db.Connection) {
	dbConn := newTestDbConnection(t)
	repos := repositories.New(dbConn)

	manager := NewManager(repos)

	return manager, dbConn
}

func asyncStartManagerAndAssertNoError(
	t *testing.T,
	manager Manager,
) *sync.WaitGroup {
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()

		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Manager panicked: %v", r)
			}
		}()

		err := manager.Start()
		assert.Nil(t, err, "Actual err: %v", err)
	}()

	time.Sleep(50 * time.Millisecond)

	return &wg
}

type mockClient struct {
	stopCalled    int
	enqueueCalled int
	enqueued      []persistence.Message
}

func (m *mockClient) Start() error {
	return nil
}

func (m *mockClient) Stop() error {
	m.stopCalled++
	return nil
}

func (m *mockClient) Enqueue(msg persistence.Message) {
	m.enqueueCalled++
	m.enqueued = append(m.enqueued, msg)
}
