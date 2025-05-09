package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/db"
	"github.com/Knoblauchpilze/chat-server/internal/service"
	"github.com/Knoblauchpilze/chat-server/pkg/clients"
	"github.com/Knoblauchpilze/chat-server/pkg/communication"
	"github.com/Knoblauchpilze/chat-server/pkg/messages"
	"github.com/Knoblauchpilze/chat-server/pkg/persistence"
	"github.com/Knoblauchpilze/chat-server/pkg/repositories"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestIT_ChatsController_PostMessageForRoom_WhenMessageHasWrongSyntax_ExpectBadRequest(t *testing.T) {
	service, dbConn, _ := newTestMessageService(t)
	defer dbConn.Close(context.Background())
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("not-a-message-dto-request"))
	ctx, rw := generateTestEchoContextFromRequest(req)
	ctx.SetParamNames("id")
	ctx.SetParamValues(uuid.NewString())

	err := postMessage(ctx, service)
	assert.Nil(t, err, "Actual err: %v", err)

	assert.Equal(t, http.StatusBadRequest, rw.Code)
	expectedBody := []byte("\"Invalid message syntax\"\n")
	assert.Equal(
		t,
		expectedBody,
		rw.Body.Bytes(),
		"Actual body: %s",
		rw.Body.String(),
	)
}

func TestIT_ChatsController_PostMessageForRoom_WhenRoomHasEmptyName_ExpectBadRequest(t *testing.T) {
	service, dbConn, _ := newTestMessageService(t)
	defer dbConn.Close(context.Background())
	requestDto := communication.MessageDtoRequest{
		Message: "",
	}

	var body bytes.Buffer
	err := json.NewEncoder(&body).Encode(requestDto)
	assert.Nil(t, err, "Actual err: %v", err)

	req := httptest.NewRequest(http.MethodPost, "/", &body)
	req.Header.Set("Content-Type", "application/json")
	ctx, rw := generateTestEchoContextFromRequest(req)
	ctx.SetParamNames("id")
	ctx.SetParamValues(uuid.NewString())

	err = postMessage(ctx, service)

	assert.Nil(t, err, "Actual err: %v", err)
	assert.Equal(t, http.StatusBadRequest, rw.Code)
	expectedBody := []byte("\"Invalid empty message\"\n")
	assert.Equal(
		t,
		expectedBody,
		rw.Body.Bytes(),
		"Actual body: %s",
		rw.Body.String(),
	)
}

func TestIT_ChatsController_PostMessageForRoom_WhenUserIsNotInRoom_ExpectBadRequest(t *testing.T) {
	service, dbConn, _ := newTestMessageService(t)
	defer dbConn.Close(context.Background())
	requestDto := communication.MessageDtoRequest{
		User:    uuid.New(),
		Room:    uuid.New(),
		Message: "hello there",
	}

	var body bytes.Buffer
	err := json.NewEncoder(&body).Encode(requestDto)
	assert.Nil(t, err, "Actual err: %v", err)

	req := httptest.NewRequest(http.MethodPost, "/", &body)
	req.Header.Set("Content-Type", "application/json")
	ctx, rw := generateTestEchoContextFromRequest(req)
	ctx.SetParamNames("id")
	ctx.SetParamValues(uuid.NewString())

	err = postMessage(ctx, service)

	assert.Nil(t, err, "Actual err: %v", err)
	assert.Equal(t, http.StatusBadRequest, rw.Code)
	expectedBody := []byte("\"User is not registered in the room\"\n")
	assert.Equal(
		t,
		expectedBody,
		rw.Body.Bytes(),
		"Actual body: %s",
		rw.Body.String(),
	)
}

func TestIT_ChatsController_PostMessageForRoom_ReturnsAccepted(t *testing.T) {
	service, dbConn, _ := newTestMessageService(t)
	defer dbConn.Close(context.Background())
	user := insertTestUser(t, dbConn)
	room := insertTestRoom(t, dbConn)
	insertUserInRoom(t, dbConn, user.Id, room.Id)

	requestDto := communication.MessageDtoRequest{
		User:    user.Id,
		Room:    room.Id,
		Message: fmt.Sprintf("%s says hello to %s", user.Name, room.Id),
	}

	var body bytes.Buffer
	err := json.NewEncoder(&body).Encode(requestDto)
	assert.Nil(t, err, "Actual err: %v", err)

	req := httptest.NewRequest(http.MethodPost, "/", &body)
	req.Header.Set("Content-Type", "application/json")
	ctx, rw := generateTestEchoContextFromRequest(req)
	ctx.SetParamNames("id")
	ctx.SetParamValues(room.Id.String())

	err = postMessage(ctx, service)

	assert.Nil(t, err, "Actual err: %v", err)
	assert.Equal(t, http.StatusAccepted, rw.Code)
}

func TestIT_ChatsController_PostMessageForRoom_SendsMessageToProcessor(t *testing.T) {
	service, dbConn, mock := newTestMessageService(t)
	defer dbConn.Close(context.Background())
	user := insertTestUser(t, dbConn)
	room := insertTestRoom(t, dbConn)
	insertUserInRoom(t, dbConn, user.Id, room.Id)

	requestDto := communication.MessageDtoRequest{
		User:    user.Id,
		Room:    room.Id,
		Message: fmt.Sprintf("%s says hello to %s", user.Name, room.Id),
	}

	var body bytes.Buffer
	err := json.NewEncoder(&body).Encode(requestDto)
	assert.Nil(t, err, "Actual err: %v", err)

	req := httptest.NewRequest(http.MethodPost, "/", &body)
	req.Header.Set("Content-Type", "application/json")
	ctx, rw := generateTestEchoContextFromRequest(req)
	ctx.SetParamNames("id")
	ctx.SetParamValues(room.Id.String())

	err = postMessage(ctx, service)

	// Wait for message to be processed
	time.Sleep(50 * time.Millisecond)

	assert.Nil(t, err, "Actual err: %v", err)
	assert.Equal(t, http.StatusAccepted, rw.Code)

	assert.Len(t, mock.enqueued, 1)
	actual := mock.enqueued[0]
	assert.Equal(t, requestDto.User, actual.ChatUser)
	assert.Equal(t, requestDto.Room, actual.Room)
	assert.Equal(t, requestDto.Message, actual.Message)
}

func TestIT_ChatsController_PostMessageForRoom_WhenRoomInRequestDoesNotMatchRoute_ExpectOverridden(t *testing.T) {
	service, dbConn, mock := newTestMessageService(t)
	defer dbConn.Close(context.Background())
	user := insertTestUser(t, dbConn)
	room := insertTestRoom(t, dbConn)
	insertUserInRoom(t, dbConn, user.Id, room.Id)

	otherId := uuid.New()
	assert.NotEqual(t, otherId, room.Id)
	requestDto := communication.MessageDtoRequest{
		User:    user.Id,
		Room:    otherId,
		Message: fmt.Sprintf("%s says hello", user.Name),
	}

	var body bytes.Buffer
	err := json.NewEncoder(&body).Encode(requestDto)
	assert.Nil(t, err, "Actual err: %v", err)

	req := httptest.NewRequest(http.MethodPost, "/", &body)
	req.Header.Set("Content-Type", "application/json")
	ctx, rw := generateTestEchoContextFromRequest(req)
	ctx.SetParamNames("id")
	ctx.SetParamValues(room.Id.String())

	err = postMessage(ctx, service)

	// Wait for message to be processed
	time.Sleep(50 * time.Millisecond)

	assert.Nil(t, err, "Actual err: %v", err)
	assert.Equal(t, http.StatusAccepted, rw.Code)

	assert.Len(t, mock.enqueued, 1)
	actual := mock.enqueued[0]
	assert.Equal(t, room.Id, actual.Room)
}

func TestIT_ChatsController_SubscribeToMessages_WhenIdHasWrongSyntax_ExpectBadRequest(t *testing.T) {
	service, dbConn, _ := newTestMessageService(t)
	defer dbConn.Close(context.Background())
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx, rw := generateTestEchoContextFromRequest(req)

	err := subscribeToMessages(ctx, service)
	assert.Nil(t, err, "Actual err: %v", err)

	assert.Equal(t, http.StatusBadRequest, rw.Code)
	expectedBody := []byte("\"Invalid id syntax\"\n")
	assert.Equal(
		t,
		expectedBody,
		rw.Body.Bytes(),
		"Actual body: %s",
		rw.Body.String(),
	)
}

func TestIT_ChatsController_SubscribeToMessage_ReceivesPostedMessage(t *testing.T) {
	dbConn := newTestDbConnection(t)
	defer dbConn.Close(context.Background())
	repos := repositories.New(dbConn)
	manager := clients.NewManager(repos)
	processor := messages.NewMessageProcessor(1, manager, repos)
	opts := service.MessageServiceOpts{
		DbConn:                 dbConn,
		Repos:                  repos,
		Processor:              processor,
		Manager:                manager,
		ClientMessageQueueSize: 1,
	}
	service := service.NewMessageService(opts)

	user1 := insertTestUser(t, dbConn)
	user2 := insertTestUser(t, dbConn)
	room := insertTestRoom(t, dbConn)
	insertUserInRoom(t, dbConn, user1.Id, room.Id)
	insertUserInRoom(t, dbConn, user2.Id, room.Id)

	requestDto := communication.MessageDtoRequest{
		User:    user2.Id,
		Room:    room.Id,
		Message: fmt.Sprintf("%s says hello to %s", user2.Name, room.Id),
	}

	wg := asyncStartProcessorAndAssertNoError(t, processor)

	// Post a message with a delay to let the user the time to subscribe
	go func() {
		time.Sleep(100 * time.Millisecond)

		var body bytes.Buffer
		err := json.NewEncoder(&body).Encode(requestDto)
		assert.Nil(t, err, "Actual err: %v", err)

		req := httptest.NewRequest(http.MethodPost, "/", &body)
		req.Header.Set("Content-Type", "application/json")
		ctx, rw := generateTestEchoContextFromRequest(req)
		ctx.SetParamNames("id")
		ctx.SetParamValues(room.Id.String())

		err = postMessage(ctx, service)
		assert.Nil(t, err, "Actual err: %v", err)
		assert.Equal(t, http.StatusAccepted, rw.Code)
	}()

	// Give the message a bit of time to be processed
	reqCtx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	req := httptest.NewRequestWithContext(reqCtx, http.MethodGet, "/", nil)
	ctx, rw := generateTestEchoContextFromRequest(req)
	ctx.SetParamNames("id")
	ctx.SetParamValues(user1.Id.String())

	err := subscribeToMessages(ctx, service)
	assert.Nil(t, err, "Actual err: %v", err)

	err = processor.Stop()
	assert.Nil(t, err, "Actual err: %v", err)
	wg.Wait()

	assert.Equal(t, http.StatusOK, rw.Code)

	// We don't know the id of the message so we need to use a regex
	// https://ihateregex.io/expr/uuid/
	const uuidPattern = "[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}"
	const timePatten = "[0-9-]+T[0-9:.]+Z"
	expectedBodyRegex := fmt.Sprintf(
		`id: %s
data: {"id":"%s","user":"%s","room":"%s","message":"%s","created_at":"%s"}

`,
		uuidPattern,
		uuidPattern,
		user2.Id.String(),
		room.Id.String(),
		requestDto.Message,
		timePatten,
	)
	actual := rw.Body.String()
	assert.Regexp(
		t,
		regexp.MustCompile(expectedBodyRegex),
		actual,
		"Actual body: %s",
		actual,
	)
}

func newTestMessageService(
	t *testing.T,
) (service.MessageService, db.Connection, *mockProcessor) {
	dbConn := newTestDbConnection(t)
	repos := repositories.New(dbConn)
	mock := &mockProcessor{}
	opts := service.MessageServiceOpts{
		DbConn:                 dbConn,
		Repos:                  repos,
		Processor:              mock,
		Manager:                nil,
		ClientMessageQueueSize: 1,
	}
	return service.NewMessageService(opts), dbConn, mock
}

func asyncStartProcessorAndAssertNoError(
	t *testing.T, processor messages.Processor,
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
		assert.Nil(t, err, "Actual err: %v", err)
	}()

	// Wait a bit for the processor to start
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
