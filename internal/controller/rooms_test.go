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
	"time"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/db"
	"github.com/Knoblauchpilze/chat-server/internal/service"
	"github.com/Knoblauchpilze/chat-server/pkg/communication"
	"github.com/Knoblauchpilze/chat-server/pkg/persistence"
	"github.com/Knoblauchpilze/chat-server/pkg/repositories"
	eassert "github.com/Knoblauchpilze/easy-assert/assert"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestIT_RoomController_CreateRoom_WhenRoomHasWrongSyntax_ExpectBadRequest(t *testing.T) {
	service, _ := newTestRoomService(t)
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
	service, _ := newTestRoomService(t)
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
	assert.Equal(t, room.Id, responseDto.Id)
	assert.Equal(t, room.Name, responseDto.Name)
	safetyMargin := 1 * time.Second
	assert.True(t, eassert.AreTimeCloserThan(room.CreatedAt, responseDto.CreatedAt, safetyMargin))
}

func TestIT_RoomController_GetRoom_WhenRoomDoesNotExist_ExpectNotFound(t *testing.T) {
	service, _ := newTestRoomService(t)

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

func TestIT_RoomController_DeleteRoom_WhenIdHasWrongSyntax_ExpectBadRequest(t *testing.T) {
	service, _ := newTestRoomService(t)
	req := httptest.NewRequest(http.MethodDelete, "/not-a-uuid", nil)
	ctx, rw := generateTestEchoContextFromRequest(req)

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
	service, _ := newTestRoomService(t)

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

	id := uuid.New()
	room := persistence.Room{
		Id:        id,
		Name:      fmt.Sprintf("my-room-%s", id),
		CreatedAt: time.Now(),
	}
	out, err := repo.Create(context.Background(), room)
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
