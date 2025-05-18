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
	"github.com/Knoblauchpilze/chat-server/pkg/repositories"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestIT_RoomController_CreateRoom_WhenRoomHasWrongSyntax_ExpectBadRequest(t *testing.T) {
	service, dbConn := newTestRoomService(t)
	defer dbConn.Close(context.Background())
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("not-a-room-dto-request"))
	ctx, rw := generateTestEchoContextFromRequest(req)

	err := createRoom(ctx, service)
	assert.Nil(t, err, "Actual err: %v", err)

	assert.Equal(t, http.StatusBadRequest, rw.Code)
	expectedBody := []byte("\"Invalid room syntax\"\n")
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
	ctx, rw := generateTestEchoContextFromRequest(req)

	err = createRoom(ctx, service)

	assert.Nil(t, err, "Actual err: %v", err)
	assert.Equal(t, http.StatusBadRequest, rw.Code)
	expectedBody := []byte("\"Invalid room name\"\n")
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
	ctx, rw := generateTestEchoContextFromRequest(req)

	err = createRoom(ctx, service)

	assert.Nil(t, err, "Actual err: %v", err)

	var responseDto communication.RoomDtoResponse
	err = json.Unmarshal(rw.Body.Bytes(), &responseDto)
	assert.Nil(t, err, "Actual err: %v", err)

	assert.Equal(t, http.StatusCreated, rw.Code)
	assert.Equal(t, requestDto.Name, responseDto.Name)
	assertRoomExists(t, dbConn, responseDto.Id)
}

func TestIT_RoomController_GetRoom(t *testing.T) {
	service, dbConn := newTestRoomService(t)
	defer dbConn.Close(context.Background())
	room := insertTestRoom(t, dbConn)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx, rw := generateTestEchoContextFromRequest(req)
	ctx.SetParamNames("id")
	ctx.SetParamValues(room.Id.String())

	err := getRoom(ctx, service)
	assert.Nil(t, err, "Actual err: %v", err)

	var responseDto communication.RoomDtoResponse
	err = json.Unmarshal(rw.Body.Bytes(), &responseDto)
	assert.Nil(t, err, "Actual err: %v", err)

	assert.Equal(t, http.StatusOK, rw.Code)
	expected := communication.ToRoomDtoResponse(room)
	assert.Equal(t, expected, responseDto)
}

func TestIT_RoomController_GetRoom_WhenRoomDoesNotExist_ExpectNotFound(t *testing.T) {
	service, dbConn := newTestRoomService(t)
	defer dbConn.Close(context.Background())

	nonExistingId := uuid.MustParse("00000000-0000-1221-0000-000000000000")
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx, rw := generateTestEchoContextFromRequest(req)
	ctx.SetParamNames("id")
	ctx.SetParamValues(nonExistingId.String())

	err := getRoom(ctx, service)
	assert.Nil(t, err, "Actual err: %v", err)

	assert.Equal(t, http.StatusNotFound, rw.Code)
	expectedBody := []byte("\"No such room\"\n")
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
	ctx, rw := generateTestEchoContextFromRequest(req)

	err := listRoom(ctx, service)
	assert.Nil(t, err, "Actual err: %v", err)

	var responseDtos []communication.RoomDtoResponse
	err = json.Unmarshal(rw.Body.Bytes(), &responseDtos)
	assert.Nil(t, err, "Actual err: %v", err)

	assert.Equal(t, http.StatusOK, rw.Code)
	expected := communication.ToRoomDtoResponse(room)
	assert.Contains(t, responseDtos, expected)
}

func TestIT_RoomController_ListUserForRoom_WhenIdHasWrongSyntax_ExpectBadRequest(t *testing.T) {
	service, dbConn := newTestRoomService(t)
	defer dbConn.Close(context.Background())
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx, rw := generateTestEchoContextFromRequest(req)
	ctx.SetParamNames("id")
	ctx.SetParamValues("not-a-uuid")

	err := listUserForRoom(ctx, service)
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
	ctx, rw := generateTestEchoContextFromRequest(req)
	ctx.SetParamNames("id")
	ctx.SetParamValues(room.Id.String())

	err := listUserForRoom(ctx, service)
	assert.Nil(t, err, "Actual err: %v", err)

	var responseDto []communication.UserDtoResponse
	err = json.Unmarshal(rw.Body.Bytes(), &responseDto)
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
	ctx, rw := generateTestEchoContextFromRequest(req)
	ctx.SetParamNames("id")
	ctx.SetParamValues(room.Id.String())

	err := listUserForRoom(ctx, service)
	assert.Nil(t, err, "Actual err: %v", err)

	var responseDto []communication.UserDtoResponse
	err = json.Unmarshal(rw.Body.Bytes(), &responseDto)
	assert.Nil(t, err, "Actual err: %v", err)

	assert.Equal(t, []communication.UserDtoResponse{}, responseDto)
}

func TestIT_RoomController_ListMessageForRoom_WhenIdHasWrongSyntax_ExpectBadRequest(t *testing.T) {
	service, dbConn := newTestRoomService(t)
	defer dbConn.Close(context.Background())
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx, rw := generateTestEchoContextFromRequest(req)

	err := listMessageForRoom(ctx, service)
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
	ctx, rw := generateTestEchoContextFromRequest(req)
	ctx.SetParamNames("id")
	ctx.SetParamValues(room.Id.String())

	err := listMessageForRoom(ctx, service)
	assert.Nil(t, err, "Actual err: %v", err)

	var responseDto []communication.MessageDtoResponse
	err = json.Unmarshal(rw.Body.Bytes(), &responseDto)
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
	ctx, rw := generateTestEchoContextFromRequest(req)
	ctx.SetParamNames("id")
	ctx.SetParamValues(room.Id.String())

	err := listMessageForRoom(ctx, service)
	assert.Nil(t, err, "Actual err: %v", err)

	var responseDto []communication.UserDtoResponse
	err = json.Unmarshal(rw.Body.Bytes(), &responseDto)
	assert.Nil(t, err, "Actual err: %v", err)

	assert.Equal(t, []communication.UserDtoResponse{}, responseDto)
}

func TestIT_RoomController_PostRegisterUserInRoom_WhenIsHasWrongSyntax_ExpectBadRequest(t *testing.T) {
	service, dbConn := newTestRoomService(t)
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

	err = postRegisterUserInRoom(ctx, service)

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

func TestIT_RoomController_PostRegisterUserInRoom_WhenRegistrationHasWrongSyntax_ExpectBadRequest(t *testing.T) {
	service, dbConn := newTestRoomService(t)
	defer dbConn.Close(context.Background())
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("not-a-registration-dto-request"))
	ctx, rw := generateTestEchoContextFromRequest(req)
	ctx.SetParamNames("id")
	ctx.SetParamValues(uuid.NewString())

	err := postRegisterUserInRoom(ctx, service)
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

func TestIT_RoomController_PostRegisterUserInRoom_WhenUserAlreadyRegistered_ExpectConflict(t *testing.T) {
	service, dbConn := newTestRoomService(t)
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

	err = postRegisterUserInRoom(ctx, service)

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

func TestIT_RoomController_PostRegisterUserInRoom(t *testing.T) {
	service, dbConn := newTestRoomService(t)
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

	err = postRegisterUserInRoom(ctx, service)

	assert.Nil(t, err, "Actual err: %v", err)

	assert.Equal(t, http.StatusNoContent, rw.Code)
	assert.Equal(t, []byte(nil), rw.Body.Bytes(), "Actual body: %s", rw.Body.String())
	assertUserRegisteredInRoom(t, dbConn, user.Id, room.Id)
}

func TestIT_RoomController_DeleteRoom_WhenIdHasWrongSyntax_ExpectBadRequest(t *testing.T) {
	service, dbConn := newTestRoomService(t)
	defer dbConn.Close(context.Background())
	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	ctx, rw := generateTestEchoContextFromRequest(req)
	ctx.SetParamNames("id")
	ctx.SetParamValues("not-a-uuid")

	err := deleteRoom(ctx, service)
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

func TestIT_RoomController_DeleteRoom(t *testing.T) {
	service, dbConn := newTestRoomService(t)
	defer dbConn.Close(context.Background())
	room := insertTestRoom(t, dbConn)

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	ctx, rw := generateTestEchoContextFromRequest(req)
	ctx.SetParamNames("id")
	ctx.SetParamValues(room.Id.String())

	err := deleteRoom(ctx, service)
	assert.Nil(t, err, "Actual err: %v", err)

	assert.Equal(t, http.StatusNoContent, rw.Code)
	assertRoomDoesNotExist(t, dbConn, room.Id)
}

func TestIT_RoomController_DeleteRoom_WhenRoomDoesNotExist_ExpectSuccess(t *testing.T) {
	service, dbConn := newTestRoomService(t)
	defer dbConn.Close(context.Background())

	nonExistingId := uuid.MustParse("00000000-0000-1221-0000-000000000000")
	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	ctx, rw := generateTestEchoContextFromRequest(req)
	ctx.SetParamNames("id")
	ctx.SetParamValues(nonExistingId.String())

	err := deleteRoom(ctx, service)
	assert.Nil(t, err, "Actual err: %v", err)

	assert.Equal(t, http.StatusNoContent, rw.Code)
}

func newTestRoomService(t *testing.T) (service.RoomService, db.Connection) {
	dbConn := newTestDbConnection(t)
	repos := repositories.New(dbConn)
	return service.NewRoomService(dbConn, repos), dbConn
}

func insertTestRoom(t *testing.T, conn db.Connection) persistence.Room {
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

func assertRoomExists(t *testing.T, dbConn db.Connection, id uuid.UUID) {
	value, err := db.QueryOne[uuid.UUID](
		context.Background(),
		dbConn,
		"SELECT id FROM room WHERE id = $1",
		id,
	)
	assert.Nil(t, err, "Actual err: %v", err)
	assert.Equal(t, id, value)
}

func assertRoomDoesNotExist(t *testing.T, conn db.Connection, id uuid.UUID) {
	value, err := db.QueryOne[int](
		context.Background(),
		conn,
		"SELECT COUNT(id) FROM room WHERE id = $1",
		id,
	)
	assert.Nil(t, err, "Actual err: %v", err)
	assert.Zero(t, value)
}
