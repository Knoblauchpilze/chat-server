package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/db"
	"github.com/Knoblauchpilze/backend-toolkit/pkg/errors"
	"github.com/Knoblauchpilze/chat-server/pkg/clients"
	"github.com/Knoblauchpilze/chat-server/pkg/communication"
	"github.com/Knoblauchpilze/chat-server/pkg/messages"
	"github.com/Knoblauchpilze/chat-server/pkg/persistence"
	"github.com/Knoblauchpilze/chat-server/pkg/repositories"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestIT_MessageService_PostMessage_SendsMessageToProcessor(t *testing.T) {
	mock := &mockProcessor{}
	service, dbConn := newTestMessageService(t, mock, nil)
	defer dbConn.Close(context.Background())
	user := insertTestUser(t, dbConn)
	room := insertTestRoom(t, dbConn)
	insertUserInRoom(t, dbConn, user.Id, room.Id)

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
	mock := &mockProcessor{}
	service, dbConn := newTestMessageService(t, mock, nil)
	defer dbConn.Close(context.Background())
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

func TestIT_MessageService_ServeClient_WhenContextTerminates_ExpectStops(t *testing.T) {
	manager := clients.NewManager()
	service, dbConn := newTestMessageService(t, nil, manager)
	defer dbConn.Close(context.Background())

	rec := httptest.NewRecorder()
	response := echo.NewResponse(rec, echo.New())

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := service.ServeClient(ctx, uuid.New(), response)

	assert.Nil(t, err, "Actual err: %v", err)
}

func TestIT_MessageService_ServeClient_WhenMessageEnqueued_ExpectClientReceivesIt(t *testing.T) {
	dbConn := newTestDbConnection(t)
	defer dbConn.Close(context.Background())
	repos := repositories.New(dbConn)
	manager := clients.NewManager()
	processor := messages.NewMessageProcessor(1, manager, repos)
	service := NewMessageService(dbConn, processor, manager)

	user1 := insertTestUser(t, dbConn)
	room := insertTestRoom(t, dbConn)
	insertUserInRoom(t, dbConn, user1.Id, room.Id)
	user2 := insertTestUser(t, dbConn)
	insertUserInRoom(t, dbConn, user2.Id, room.Id)

	rec := httptest.NewRecorder()
	response := echo.NewResponse(rec, echo.New())

	ctx, cancel := context.WithCancel(context.Background())

	msg := persistence.Message{
		Id:       uuid.New(),
		ChatUser: user2.Id,
		Room:     room.Id,
		Message:  "Hello",
	}
	processor.Enqueue(msg)

	wgService := asyncServeClientAndAssertNoError(t, service, ctx, user1.Id, response)
	// Wait for the client to be registered
	time.Sleep(50 * time.Millisecond)
	wgProcessor := asyncStartMessageProcessorAndAssertNoError(t, processor)

	// Wait for the message to be processed
	time.Sleep(50 * time.Millisecond)

	cancel()
	err := processor.Stop()
	assert.Nil(t, err, "Actual err: %v", err)

	wgProcessor.Wait()
	wgService.Wait()

	assert.Equal(t, http.StatusOK, rec.Code)
	body, err := io.ReadAll(rec.Body)
	assert.Nil(t, err, "Actual err: %v", err)

	received := communication.ToMessageDtoResponse(msg)
	msgAsJson, err := json.Marshal(received)
	assert.Nil(t, err, "Actual err: %v", err)

	expected := fmt.Sprintf(
		`id: %s
data: %s

`,
		msg.Id.String(),
		msgAsJson,
	)
	assert.Equal(
		t,
		[]byte(expected),
		body,
		"Expected %s, got: %s",
		string(expected),
		string(body),
	)
}

func TestIT_MessageService_ServeClient_WhenMessageFromClientReceived_ExpectClientDoesNotReceiveIt(t *testing.T) {
	dbConn := newTestDbConnection(t)
	defer dbConn.Close(context.Background())
	repos := repositories.New(dbConn)
	manager := clients.NewManager()
	processor := messages.NewMessageProcessor(1, manager, repos)
	service := NewMessageService(dbConn, processor, manager)

	user := insertTestUser(t, dbConn)
	room := insertTestRoom(t, dbConn)
	insertUserInRoom(t, dbConn, user.Id, room.Id)

	rec := httptest.NewRecorder()
	response := echo.NewResponse(rec, echo.New())

	ctx, cancel := context.WithCancel(context.Background())

	msg := persistence.Message{
		Id:       uuid.New(),
		ChatUser: user.Id,
		Room:     room.Id,
		Message:  "Hello",
	}
	processor.Enqueue(msg)

	wgService := asyncServeClientAndAssertNoError(t, service, ctx, user.Id, response)
	// Wait for the client to be registered
	time.Sleep(50 * time.Millisecond)
	wgProcessor := asyncStartMessageProcessorAndAssertNoError(t, processor)

	// Wait for the message to be processed
	time.Sleep(50 * time.Millisecond)

	cancel()
	err := processor.Stop()
	assert.Nil(t, err, "Actual err: %v", err)

	wgProcessor.Wait()
	wgService.Wait()

	assert.Equal(t, http.StatusOK, rec.Code)
	body, err := io.ReadAll(rec.Body)
	assert.Nil(t, err, "Actual err: %v", err)

	assert.Len(t, body, 0, "Received unexpected body: %s", string(body))
}

func newTestMessageService(
	t *testing.T,
	processor messages.Processor,
	manager clients.Manager,
) (MessageService, db.Connection) {
	dbConn := newTestDbConnection(t)
	return NewMessageService(dbConn, processor, manager), dbConn
}

func asyncStartMessageProcessorAndAssertNoError(
	t *testing.T,
	processor messages.Processor,
) *sync.WaitGroup {
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		defer func() {
			if r := recover(); r != nil {
				assert.Fail(t, "Processor panicked: %v", r)
			}
		}()

		err := processor.Start()
		assert.Nil(t, err, "Actual err: %v", err)
	}()

	time.Sleep(50 * time.Millisecond)

	return &wg
}

func asyncServeClientAndAssertNoError(
	t *testing.T,
	service MessageService,
	ctx context.Context,
	client uuid.UUID,
	response *echo.Response,
) *sync.WaitGroup {
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		defer func() {
			if r := recover(); r != nil {
				assert.Fail(t, "Serve client panicked: %v", r)
			}
		}()

		err := service.ServeClient(ctx, client, response)
		assert.Nil(t, err, "Actual err: %v", err)
	}()

	time.Sleep(50 * time.Millisecond)

	return &wg
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

func (m *mockProcessor) Enqueue(msg persistence.Message) {
	m.enqueued = append(m.enqueued, msg)
}
