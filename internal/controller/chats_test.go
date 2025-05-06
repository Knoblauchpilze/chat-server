package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/db"
	"github.com/Knoblauchpilze/chat-server/internal/service"
	"github.com/Knoblauchpilze/chat-server/pkg/communication"
	"github.com/Knoblauchpilze/chat-server/pkg/persistence"
	"github.com/stretchr/testify/assert"
)

func TestIT_ChatsController_PostMessageForRoom_WhenMessageHasWrongSyntax_ExpectBadRequest(t *testing.T) {
	service, dbConn := newTestMessageService(t)
	defer dbConn.Close(context.Background())
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("not-a-message-dto-request"))
	ctx, rw := generateTestEchoContextFromRequest(req)

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
	service, dbConn := newTestMessageService(t)
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

func TestIT_ChatsController_PostMessageForRoom(t *testing.T) {
	service, dbConn := newTestMessageService(t)
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

	err = postMessage(ctx, service)

	assert.Nil(t, err, "Actual err: %v", err)
	assert.Equal(t, http.StatusAccepted, rw.Code)
}

func TestIT_ChatsController_SubscribeToMessages_WhenIdHasWrongSyntax_ExpectBadRequest(t *testing.T) {
	service, dbConn := newTestMessageService(t)
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

func newTestMessageService(
	t *testing.T,
) (service.MessageService, db.Connection) {
	dbConn := newTestDbConnection(t)
	// TODO: Correctly setup the client manager
	return service.NewMessageService(dbConn, &mockProcessor{}, nil), dbConn
}

type mockProcessor struct{}

func (m *mockProcessor) Start() error {
	return nil
}

func (m *mockProcessor) Stop() error {
	return nil
}

func (m *mockProcessor) Enqueue(_ persistence.Message) {}
