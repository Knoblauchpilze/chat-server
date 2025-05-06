package clients

import (
	"sync"
	"testing"
	"time"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/errors"
	"github.com/Knoblauchpilze/chat-server/pkg/persistence"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestUnit_Manager_WhenClientAlreadyRegistered_ExpectError(t *testing.T) {
	manager := NewManager()
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
	manager := NewManager()
	id := uuid.New()
	mock := &mockClient{}

	err := manager.OnConnect(id, mock)
	assert.Nil(t, err, "Actual err: %v", err)

	asyncStartManagerAndAssertNoError(t, manager)

	err = manager.Stop()
	assert.Nil(t, err, "Actual err: %v", err)

	assert.Equal(t, 1, mock.stopCalled)
}

func TestUnit_Manager_WhenBroadcast_ExpectMessageReceived(t *testing.T) {
	manager := NewManager()
	id := uuid.New()
	mock := &mockClient{}

	err := manager.OnConnect(id, mock)
	assert.Nil(t, err, "Actual err: %v", err)

	msg := persistence.Message{
		Id:        uuid.New(),
		ChatUser:  uuid.New(),
		Room:      uuid.New(),
		Message:   "Hello",
		CreatedAt: time.Date(2025, 5, 5, 21, 44, 20, 0, time.UTC),
	}
	manager.Broadcast(msg)

	assert.Equal(t, 1, mock.enqueueCalled)
	expected := []persistence.Message{msg}
	assert.Equal(t, expected, mock.enqueued, 1)
}

func TestUnit_Manager_WhenBroadcastAfterDisconnect_ExpectNoMessageReceived(t *testing.T) {
	manager := NewManager()
	id := uuid.New()
	mock := &mockClient{}

	err := manager.OnConnect(id, mock)
	assert.Nil(t, err, "Actual err: %v", err)
	manager.OnDisconnect(id)

	msg := persistence.Message{
		Id:        uuid.New(),
		ChatUser:  uuid.New(),
		Room:      uuid.New(),
		Message:   "Hello",
		CreatedAt: time.Date(2025, 5, 5, 21, 44, 20, 0, time.UTC),
	}
	manager.Broadcast(msg)

	assert.Equal(t, 0, mock.enqueueCalled)
}

func TestUnit_Manager_WhenBroadcastExceptToClient_ExpectNoMessageReceived(t *testing.T) {
	manager := NewManager()
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
	manager.BroadcastExcept(clientId2, msg)

	assert.Equal(t, 1, mock1.enqueueCalled)
	expected := []persistence.Message{msg}
	assert.Equal(t, expected, mock1.enqueued, 1)

	assert.Equal(t, 0, mock2.enqueueCalled)
}

func TestUnit_Manager_WhenSendingMessageToSpecificClient_ExpectMessageReceived(t *testing.T) {
	manager := NewManager()
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
