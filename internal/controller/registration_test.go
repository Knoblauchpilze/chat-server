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
	"github.com/gin-gonic/gin"
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
	ctx, rw := generateTestGinContextFromRequest(req)
	ctx.Params = gin.Params{{Key: "id", Value: "not-a-uuid"}}

	addUserInRoom(ctx, service)

	assert.Equal(t, http.StatusBadRequest, ctx.Writer.Status())
	expectedBody := []byte("\"Invalid id syntax\"")
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
	ctx, rw := generateTestGinContextFromRequest(req)
	ctx.Params = gin.Params{{Key: "id", Value: uuid.NewString()}}

	addUserInRoom(ctx, service)

	assert.Equal(t, http.StatusBadRequest, ctx.Writer.Status())
	expectedBody := []byte("\"Invalid registration syntax\"")
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
	ctx, rw := generateTestGinContextFromRequest(req)
	ctx.Params = gin.Params{{Key: "id", Value: room.Id.String()}}

	addUserInRoom(ctx, service)

	assert.Equal(t, http.StatusConflict, ctx.Writer.Status())
	assert.Equal(
		t,
		[]byte("\"User already registered in room\""),
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
	ctx, rw := generateTestGinContextFromRequest(req)
	ctx.Params = gin.Params{{Key: "id", Value: room.Id.String()}}

	addUserInRoom(ctx, service)

	assert.Equal(t, http.StatusNoContent, ctx.Writer.Status())
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
	ctx, _ := generateTestGinContextFromRequest(req)
	ctx.Params = gin.Params{
		{
			Key:   "id",
			Value: room.Id.String(),
		},
		{
			Key:   "user",
			Value: user.Id.String(),
		},
	}

	deleteUserFromRoom(ctx, service)

	assert.Equal(t, http.StatusNoContent, ctx.Writer.Status())
	assertUserNotRegisteredInRoom(t, dbConn, user.Id, room.Id)
}

func newTestRegistrationService(t *testing.T) (service.RegistrationService, *db.Connection) {
	dbConn := newTestDbConnection(t)
	repos := repositories.New(dbConn)
	return service.NewRegistrationService(dbConn, repos), dbConn
}
