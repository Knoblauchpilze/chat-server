package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/db"
	"github.com/Knoblauchpilze/chat-server/internal/service"
	"github.com/Knoblauchpilze/chat-server/pkg/communication"
	"github.com/Knoblauchpilze/chat-server/pkg/repositories"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestIT_RegistrationController_AddUserInRoom_WhenIsHasWrongSyntax_ExpectBadRequest(t *testing.T) {
	service, dbConn := newTestRegistrationService(t)
	defer dbConn.Close(context.Background())
	requestDto := communication.RoomRegistrationDtoRequest{
		User: uuid.New(),
	}

	var body bytes.Buffer
	err := json.NewEncoder(&body).Encode(requestDto)
	assert.Nil(t, err, "Actual err: %v", err)

	req := httptest.NewRequest(http.MethodPost, "/", &body)
	req.Header.Set("Content-Type", "application/json")
	ctx, rw := generateTestEchoContextFromRequest(req)
	ctx.SetParamNames("id")
	ctx.SetParamValues("not-a-uuid")

	err = addUserInRoom(ctx, service)

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

func TestIT_RegistrationController_AddUserInRoom_WhenRegistrationHasWrongSyntax_ExpectBadRequest(t *testing.T) {
	service, dbConn := newTestRegistrationService(t)
	defer dbConn.Close(context.Background())
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("not-a-registration-dto-request"))
	ctx, rw := generateTestEchoContextFromRequest(req)
	ctx.SetParamNames("id")
	ctx.SetParamValues(uuid.NewString())

	err := addUserInRoom(ctx, service)
	assert.Nil(t, err, "Actual err: %v", err)

	assert.Equal(t, http.StatusBadRequest, rw.Code)
	expectedBody := []byte("\"Invalid registration syntax\"\n")
	assert.Equal(
		t,
		expectedBody,
		rw.Body.Bytes(),
		"Actual body: %s",
		rw.Body.String(),
	)
}

func TestIT_RegistrationController_AddUserInRoom_WhenUserAlreadyRegistered_ExpectConflict(t *testing.T) {
	service, dbConn := newTestRegistrationService(t)
	defer dbConn.Close(context.Background())
	user := insertTestUser(t, dbConn)
	room := insertTestRoom(t, dbConn)
	registerUserInRoom(t, dbConn, user.Id, room.Id)
	requestDto := communication.RoomRegistrationDtoRequest{
		User: user.Id,
	}

	var body bytes.Buffer
	err := json.NewEncoder(&body).Encode(requestDto)
	assert.Nil(t, err, "Actual err: %v", err)

	req := httptest.NewRequest(http.MethodPost, "/", &body)
	req.Header.Set("Content-Type", "application/json")
	ctx, rw := generateTestEchoContextFromRequest(req)
	ctx.SetParamNames("id")
	ctx.SetParamValues(room.Id.String())

	err = addUserInRoom(ctx, service)

	assert.Nil(t, err, "Actual err: %v", err)

	assert.Equal(t, http.StatusConflict, rw.Code)
	assert.Equal(
		t,
		[]byte("\"User already registered in room\"\n"),
		rw.Body.Bytes(),
		"Actual body: %s",
		rw.Body.String(),
	)
}

func TestIT_RegistrationController_AddUserInRoom(t *testing.T) {
	service, dbConn := newTestRegistrationService(t)
	defer dbConn.Close(context.Background())
	user := insertTestUser(t, dbConn)
	room := insertTestRoom(t, dbConn)
	requestDto := communication.RoomRegistrationDtoRequest{
		User: user.Id,
	}

	var body bytes.Buffer
	err := json.NewEncoder(&body).Encode(requestDto)
	assert.Nil(t, err, "Actual err: %v", err)

	req := httptest.NewRequest(http.MethodPost, "/", &body)
	req.Header.Set("Content-Type", "application/json")
	ctx, rw := generateTestEchoContextFromRequest(req)
	ctx.SetParamNames("id")
	ctx.SetParamValues(room.Id.String())

	err = addUserInRoom(ctx, service)

	assert.Nil(t, err, "Actual err: %v", err)

	assert.Equal(t, http.StatusNoContent, rw.Code)
	assert.Equal(t, []byte(nil), rw.Body.Bytes(), "Actual body: %s", rw.Body.String())
	assertUserRegisteredInRoom(t, dbConn, user.Id, room.Id)
}

func TestIT_RegistrationController_DeleteUserFromRoom(t *testing.T) {
	service, dbConn := newTestRegistrationService(t)
	defer dbConn.Close(context.Background())
	user := insertTestUser(t, dbConn)
	room := insertTestRoom(t, dbConn)
	registerUserInRoom(t, dbConn, user.Id, room.Id)

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	ctx, rw := generateTestEchoContextFromRequest(req)
	ctx.SetParamNames("room", "user")
	ctx.SetParamValues(room.Id.String(), user.Id.String())

	err := deleteUserFromRoom(ctx, service)

	assert.Nil(t, err, "Actual err: %v", err)

	assert.Equal(t, http.StatusNoContent, rw.Code)
	assertUserNotRegisteredInRoom(t, dbConn, user.Id, room.Id)
}

func newTestRegistrationService(t *testing.T) (service.RegistrationService, db.Connection) {
	dbConn := newTestDbConnection(t)
	repos := repositories.New(dbConn)
	return service.NewRegistrationService(dbConn, repos), dbConn
}
