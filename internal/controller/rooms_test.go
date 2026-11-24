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
	"github.com/Knoblauchpilze/backend-toolkit/pkg/middleware"
	"github.com/Knoblauchpilze/chat-server/internal/service"
	"github.com/Knoblauchpilze/chat-server/pkg/communication"
	"github.com/Knoblauchpilze/chat-server/pkg/persistence"
	"github.com/Knoblauchpilze/chat-server/pkg/repositories"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestUnit_RoomController_CreateRoom_WhenInvalidJson_ExpectErrorEnvelope(t *testing.T) {
	for _, body := range []string{"{", `{"name": 42}`} {
		t.Run(body, func(t *testing.T) {
			router := gin.New()
			router.Use(middleware.ResponseEnvelope())
			route := RoomEndpoints(nil)[0]
			router.Handle(route.Method(), route.Path(), route.Handler())

			req := httptest.NewRequest(http.MethodPost, "/rooms", strings.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			rw := httptest.NewRecorder()
			router.ServeHTTP(rw, req)

			assert.Equal(t, http.StatusBadRequest, rw.Code)
			assert.JSONEq(t, `{
				"request_id": "unknown",
				"status": "ERROR",
				"status_code": 400,
				"details": "Invalid room syntax"
			}`, rw.Body.String())
		})
	}
}

func TestIT_RoomController_CreateRoom_WhenRoomHasWrongSyntax_ExpectBadRequest(t *testing.T) {
	service, dbConn := newTestRoomService(t)
	defer dbConn.Close(context.Background())
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("not-a-room-dto-request"))
	ctx, rw := generateTestGinContextFromRequest(req)

	createRoom(ctx, service)

	assert.Equal(t, http.StatusBadRequest, ctx.Writer.Status())
	expectedBody := []byte("\"Invalid room syntax\"")
	assert.Equal(
		t,
		expectedBody,
		rw.Body.Bytes(),
		"Actual body: %s",
		rw.Body.String(),
	)
}

func TestIT_RoomController_CreateRoom_WhenRoomHasEmptyName_ExpectBadRequest(t *testing.T) {
	service, dbConn := newTestRoomService(t)
	defer dbConn.Close(context.Background())
	requestDto := communication.RoomDtoRequest{
		Name: "",
	}

	var body bytes.Buffer
	err := json.NewEncoder(&body).Encode(requestDto)
	assert.Nil(t, err, "Actual err: %v", err)

	req := httptest.NewRequest(http.MethodPost, "/", &body)
	req.Header.Set("Content-Type", "application/json")
	ctx, rw := generateTestGinContextFromRequest(req)

	createRoom(ctx, service)
	assert.Equal(t, http.StatusBadRequest, ctx.Writer.Status())
	expectedBody := []byte("\"Invalid room name\"")
	assert.Equal(
		t,
		expectedBody,
		rw.Body.Bytes(),
		"Actual body: %s",
		rw.Body.String(),
	)
}

func TestIT_RoomController_CreateRoom(t *testing.T) {
	service, dbConn := newTestRoomService(t)
	defer dbConn.Close(context.Background())
	requestDto := communication.RoomDtoRequest{
		Name: fmt.Sprintf("my-room-%s", uuid.NewString()),
	}

	var body bytes.Buffer
	err := json.NewEncoder(&body).Encode(requestDto)
	assert.Nil(t, err, "Actual err: %v", err)

	req := httptest.NewRequest(http.MethodPost, "/", &body)
	req.Header.Set("Content-Type", "application/json")
	ctx, rw := generateTestGinContextFromRequest(req)

	createRoom(ctx, service)

	var responseDto communication.RoomDtoResponse
	err = json.Unmarshal(rw.Body.Bytes(), &responseDto)
	assert.Nil(t, err, "Actual err: %v", err)

	assert.Equal(t, http.StatusCreated, ctx.Writer.Status())
	assert.Equal(t, requestDto.Name, responseDto.Name)
	assertRoomExists(t, dbConn, responseDto.Id)
}

func TestIT_RoomController_GetRoom(t *testing.T) {
	service, dbConn := newTestRoomService(t)
	defer dbConn.Close(context.Background())
	room := insertTestRoom(t, dbConn)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx, rw := generateTestGinContextFromRequest(req)
	ctx.Params = gin.Params{{Key: "id", Value: room.Id.String()}}

	getRoom(ctx, service)

	var responseDto communication.RoomDtoResponse
	err := json.Unmarshal(rw.Body.Bytes(), &responseDto)
	assert.Nil(t, err, "Actual err: %v", err)

	assert.Equal(t, http.StatusOK, ctx.Writer.Status())
	expected := communication.ToRoomDtoResponse(room)
	assert.Equal(t, expected, responseDto)
}

func TestIT_RoomController_GetRoom_WhenRoomDoesNotExist_ExpectNotFound(t *testing.T) {
	service, dbConn := newTestRoomService(t)
	defer dbConn.Close(context.Background())

	nonExistingId := uuid.MustParse("00000000-0000-1221-0000-000000000000")
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx, rw := generateTestGinContextFromRequest(req)
	ctx.Params = gin.Params{{Key: "id", Value: nonExistingId.String()}}

	getRoom(ctx, service)

	assert.Equal(t, http.StatusNotFound, ctx.Writer.Status())
	expectedBody := []byte("\"No such room\"")
	assert.Equal(
		t,
		expectedBody,
		rw.Body.Bytes(),
		"Actual body: %s",
		rw.Body.String(),
	)
}

func TestIT_RoomController_ListROom(t *testing.T) {
	service, dbConn := newTestRoomService(t)
	defer dbConn.Close(context.Background())
	room := insertTestRoom(t, dbConn)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx, rw := generateTestGinContextFromRequest(req)

	listRoom(ctx, service)

	var responseDtos []communication.RoomDtoResponse
	err := json.Unmarshal(rw.Body.Bytes(), &responseDtos)
	assert.Nil(t, err, "Actual err: %v", err)

	assert.Equal(t, http.StatusOK, ctx.Writer.Status())
	expected := communication.ToRoomDtoResponse(room)
	assert.Contains(t, responseDtos, expected)
}

func TestIT_RoomController_ListUserForRoom_WhenIdHasWrongSyntax_ExpectBadRequest(t *testing.T) {
	service, dbConn := newTestRoomService(t)
	defer dbConn.Close(context.Background())
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx, rw := generateTestGinContextFromRequest(req)
	ctx.Params = gin.Params{{Key: "id", Value: "not-a-uuid"}}

	listUserForRoom(ctx, service)

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

func TestIT_RoomController_ListUserForRoom(t *testing.T) {
	service, dbConn := newTestRoomService(t)
	defer dbConn.Close(context.Background())
	user1 := insertTestUser(t, dbConn)
	user2 := insertTestUser(t, dbConn)
	insertTestUser(t, dbConn)

	room := insertTestRoom(t, dbConn)
	registerUserInRoom(t, dbConn, user1.Id, room.Id)
	registerUserInRoom(t, dbConn, user2.Id, room.Id)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx, rw := generateTestGinContextFromRequest(req)
	ctx.Params = gin.Params{{Key: "id", Value: room.Id.String()}}

	listUserForRoom(ctx, service)

	var responseDto []communication.UserDtoResponse
	err := json.Unmarshal(rw.Body.Bytes(), &responseDto)
	assert.Nil(t, err, "Actual err: %v", err)

	expected := []communication.UserDtoResponse{
		communication.ToUserDtoResponse(user1),
		communication.ToUserDtoResponse(user2),
	}
	assert.ElementsMatch(t, expected, responseDto)
}

func TestIT_RoomController_ListUserForRoom_WhenNoUserInRoom_ExpectEmptySlice(t *testing.T) {
	service, dbConn := newTestRoomService(t)
	defer dbConn.Close(context.Background())
	room := insertTestRoom(t, dbConn)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx, rw := generateTestGinContextFromRequest(req)
	ctx.Params = gin.Params{{Key: "id", Value: room.Id.String()}}

	listUserForRoom(ctx, service)

	var responseDto []communication.UserDtoResponse
	err := json.Unmarshal(rw.Body.Bytes(), &responseDto)
	assert.Nil(t, err, "Actual err: %v", err)

	assert.Equal(t, []communication.UserDtoResponse{}, responseDto)
}

func TestIT_RoomController_ListMessageForRoom_WhenIdHasWrongSyntax_ExpectBadRequest(t *testing.T) {
	service, dbConn := newTestRoomService(t)
	defer dbConn.Close(context.Background())
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx, rw := generateTestGinContextFromRequest(req)

	listMessageForRoom(ctx, service)

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

func TestIT_RoomController_ListMessageForRoom(t *testing.T) {
	service, dbConn := newTestRoomService(t)
	defer dbConn.Close(context.Background())
	room := insertTestRoom(t, dbConn)

	user1 := insertTestUser(t, dbConn)
	user2 := insertTestUser(t, dbConn)
	registerUserInRoom(t, dbConn, user1.Id, room.Id)
	registerUserInRoom(t, dbConn, user2.Id, room.Id)

	msg1 := insertTestMessage(t, dbConn, user1.Id, room.Id)
	msg2 := insertTestMessage(t, dbConn, user2.Id, room.Id)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx, rw := generateTestGinContextFromRequest(req)
	ctx.Params = gin.Params{{Key: "id", Value: room.Id.String()}}

	listMessageForRoom(ctx, service)

	var responseDto []communication.MessageDtoResponse
	err := json.Unmarshal(rw.Body.Bytes(), &responseDto)
	assert.Nil(t, err, "Actual err: %v", err)

	expected := []communication.MessageDtoResponse{
		communication.ToMessageDtoResponse(msg1),
		communication.ToMessageDtoResponse(msg2),
	}
	assert.ElementsMatch(t, expected, responseDto)
}

func TestIT_RoomController_ListMessageForRoom_WhenNoMessageInRoom_ExpectEmptySlice(t *testing.T) {
	service, dbConn := newTestRoomService(t)
	defer dbConn.Close(context.Background())
	room := insertTestRoom(t, dbConn)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx, rw := generateTestGinContextFromRequest(req)
	ctx.Params = gin.Params{{Key: "id", Value: room.Id.String()}}

	listMessageForRoom(ctx, service)

	var responseDto []communication.UserDtoResponse
	err := json.Unmarshal(rw.Body.Bytes(), &responseDto)
	assert.Nil(t, err, "Actual err: %v", err)

	assert.Equal(t, []communication.UserDtoResponse{}, responseDto)
}

func TestIT_RoomController_DeleteRoom_WhenIdHasWrongSyntax_ExpectBadRequest(t *testing.T) {
	service, dbConn := newTestRoomService(t)
	defer dbConn.Close(context.Background())
	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	ctx, rw := generateTestGinContextFromRequest(req)
	ctx.Params = gin.Params{{Key: "id", Value: "not-a-uuid"}}

	deleteRoom(ctx, service)

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

func TestIT_RoomController_DeleteRoom(t *testing.T) {
	service, dbConn := newTestRoomService(t)
	defer dbConn.Close(context.Background())
	room := insertTestRoom(t, dbConn)

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	ctx, _ := generateTestGinContextFromRequest(req)
	ctx.Params = gin.Params{{Key: "id", Value: room.Id.String()}}

	deleteRoom(ctx, service)

	assert.Equal(t, http.StatusNoContent, ctx.Writer.Status())
	assertRoomDoesNotExist(t, dbConn, room.Id)
}

func TestIT_RoomController_DeleteRoom_WhenRoomDoesNotExist_ExpectSuccess(t *testing.T) {
	service, dbConn := newTestRoomService(t)
	defer dbConn.Close(context.Background())

	nonExistingId := uuid.MustParse("00000000-0000-1221-0000-000000000000")
	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	ctx, _ := generateTestGinContextFromRequest(req)
	ctx.Params = gin.Params{{Key: "id", Value: nonExistingId.String()}}

	deleteRoom(ctx, service)

	assert.Equal(t, http.StatusNoContent, ctx.Writer.Status())
}

func newTestRoomService(t *testing.T) (service.RoomService, *db.Connection) {
	dbConn := newTestDbConnection(t)
	repos := repositories.New(dbConn)
	return service.NewRoomService(dbConn, repos), dbConn
}

func insertTestRoom(t *testing.T, conn *db.Connection) persistence.Room {
	repo := repositories.NewRoomRepository(conn)

	tx, err := conn.BeginTx(context.Background())
	assert.Nil(t, err, "Actual err: %v", err)

	id := uuid.New()
	room := persistence.Room{
		Id:   id,
		Name: fmt.Sprintf("my-room-%s", id),
	}
	out, err := repo.Create(context.Background(), tx, room)
	tx.Close(context.Background())
	assert.Nil(t, err, "Actual err: %v", err)

	assertRoomExists(t, conn, out.Id)

	return out
}

func assertRoomExists(t *testing.T, dbConn *db.Connection, id uuid.UUID) {
	value, err := db.QueryOne[uuid.UUID](
		context.Background(),
		dbConn,
		"SELECT id FROM room WHERE id = $1",
		id,
	)
	assert.Nil(t, err, "Actual err: %v", err)
	assert.Equal(t, id, value)
}

func assertRoomDoesNotExist(t *testing.T, conn *db.Connection, id uuid.UUID) {
	value, err := db.QueryOne[int](
		context.Background(),
		conn,
		"SELECT COUNT(id) FROM room WHERE id = $1",
		id,
	)
	assert.Nil(t, err, "Actual err: %v", err)
	assert.Zero(t, value)
}
