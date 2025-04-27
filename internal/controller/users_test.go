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

func TestIT_UserController_CreateUser_WhenUserHasWrongSyntax_ExpectBadRequest(t *testing.T) {
	service, dbConn := newTestUserService(t)
	defer dbConn.Close(context.Background())
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("not-a-user-dto-request"))
	ctx, rw := generateTestEchoContextFromRequest(req)

	err := createUser(ctx, service)
	assert.Nil(t, err, "Actual err: %v", err)

	assert.Equal(t, http.StatusBadRequest, rw.Code)
	expectedBody := []byte("\"Invalid user syntax\"\n")
	assert.Equal(
		t,
		expectedBody,
		rw.Body.Bytes(),
		"Actual body: %s",
		rw.Body.String(),
	)
}

func TestIT_UserController_CreateUser_WhenUserHasEmptyName_ExpectBadRequest(t *testing.T) {
	service, dbConn := newTestUserService(t)
	defer dbConn.Close(context.Background())
	requestDto := communication.UserDtoRequest{
		Name:    "",
		ApiUser: uuid.New(),
	}

	var body bytes.Buffer
	err := json.NewEncoder(&body).Encode(requestDto)
	assert.Nil(t, err, "Actual err: %v", err)

	req := httptest.NewRequest(http.MethodPost, "/", &body)
	req.Header.Set("Content-Type", "application/json")
	ctx, rw := generateTestEchoContextFromRequest(req)

	err = createUser(ctx, service)

	assert.Nil(t, err, "Actual err: %v", err)
	assert.Equal(t, http.StatusBadRequest, rw.Code)
	expectedBody := []byte("\"Invalid user name\"\n")
	assert.Equal(
		t,
		expectedBody,
		rw.Body.Bytes(),
		"Actual body: %s",
		rw.Body.String(),
	)
}

func TestIT_UserController_CreateUser(t *testing.T) {
	service, dbConn := newTestUserService(t)
	defer dbConn.Close(context.Background())
	requestDto := communication.UserDtoRequest{
		Name:    fmt.Sprintf("my-user-%s", uuid.NewString()),
		ApiUser: uuid.New(),
	}

	var body bytes.Buffer
	err := json.NewEncoder(&body).Encode(requestDto)
	assert.Nil(t, err, "Actual err: %v", err)

	req := httptest.NewRequest(http.MethodPost, "/", &body)
	req.Header.Set("Content-Type", "application/json")
	ctx, rw := generateTestEchoContextFromRequest(req)

	err = createUser(ctx, service)

	assert.Nil(t, err, "Actual err: %v", err)

	var responseDto communication.UserDtoResponse
	err = json.Unmarshal(rw.Body.Bytes(), &responseDto)
	assert.Nil(t, err, "Actual err: %v", err)

	assert.Equal(t, http.StatusCreated, rw.Code)
	assert.Equal(t, requestDto.Name, responseDto.Name)
	assertUserExists(t, dbConn, responseDto.Id)
}

func TestIT_UserController_GetUser(t *testing.T) {
	service, dbConn := newTestUserService(t)
	defer dbConn.Close(context.Background())
	user := insertTestUser(t, dbConn)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx, rw := generateTestEchoContextFromRequest(req)
	ctx.SetParamNames("id")
	ctx.SetParamValues(user.Id.String())

	err := getUser(ctx, service)
	assert.Nil(t, err, "Actual err: %v", err)

	var responseDto communication.UserDtoResponse
	err = json.Unmarshal(rw.Body.Bytes(), &responseDto)
	assert.Nil(t, err, "Actual err: %v", err)

	assert.Equal(t, http.StatusOK, rw.Code)
	expected := communication.ToUserDtoResponse(user)
	assert.Equal(t, expected, responseDto)
}

func TestIT_UserController_GetUser_WhenUserDoesNotExist_ExpectNotFound(t *testing.T) {
	service, dbConn := newTestUserService(t)
	defer dbConn.Close(context.Background())

	nonExistingId := uuid.MustParse("00000000-0000-1221-0000-000000000000")
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx, rw := generateTestEchoContextFromRequest(req)
	ctx.SetParamNames("id")
	ctx.SetParamValues(nonExistingId.String())

	err := getUser(ctx, service)
	assert.Nil(t, err, "Actual err: %v", err)

	assert.Equal(t, http.StatusNotFound, rw.Code)
	expectedBody := []byte("\"No such user\"\n")
	assert.Equal(
		t,
		expectedBody,
		rw.Body.Bytes(),
		"Actual body: %s",
		rw.Body.String(),
	)
}

func TestIT_UserController_ListUsers_WhenNoNameProvided_ExpectBadRequest(t *testing.T) {
	service, dbConn := newTestUserService(t)
	defer dbConn.Close(context.Background())
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx, rw := generateTestEchoContextFromRequest(req)

	err := listUsers(ctx, service)
	assert.Nil(t, err, "Actual err: %v", err)

	assert.Equal(t, http.StatusBadRequest, rw.Code)
	expectedBody := []byte("\"Please provide a user name as filtering parameter\"\n")
	assert.Equal(
		t,
		expectedBody,
		rw.Body.Bytes(),
		"Actual body: %s",
		rw.Body.String(),
	)
}

func TestIT_UserController_ListUsers(t *testing.T) {
	service, dbConn := newTestUserService(t)
	defer dbConn.Close(context.Background())
	user := insertTestUser(t, dbConn)
	insertTestUser(t, dbConn)

	req := generateTestRequestWithQueryParam(http.MethodGet, "name", user.Name)
	ctx, rw := generateTestEchoContextFromRequest(req)

	err := listUsers(ctx, service)
	assert.Nil(t, err, "Actual err: %v", err)

	var responseDto []communication.UserDtoResponse
	err = json.Unmarshal(rw.Body.Bytes(), &responseDto)
	assert.Nil(t, err, "Actual err: %v", err)

	expected := []communication.UserDtoResponse{
		communication.ToUserDtoResponse(user),
	}
	assert.ElementsMatch(t, expected, responseDto)
}

func TestIT_UserController_ListUsers_WhenNameHasSpecialCharacters_ExpectSuccess(t *testing.T) {
	service, dbConn := newTestUserService(t)
	defer dbConn.Close(context.Background())
	name := fmt.Sprintf("my special /chars\\ user 😊 %s", uuid.NewString())
	user := insertTestUserWithName(t, dbConn, name)

	req := generateTestRequestWithQueryParam(http.MethodGet, "name", user.Name)
	ctx, rw := generateTestEchoContextFromRequest(req)

	err := listUsers(ctx, service)
	assert.Nil(t, err, "Actual err: %v", err)

	var responseDto []communication.UserDtoResponse
	err = json.Unmarshal(rw.Body.Bytes(), &responseDto)
	assert.Nil(t, err, "Actual err: %v", err)

	expected := []communication.UserDtoResponse{
		communication.ToUserDtoResponse(user),
	}

	assert.ElementsMatch(t, expected, responseDto)
}

func TestIT_UserController_ListForUser_WhenIdHasWrongSyntax_ExpectBadRequest(t *testing.T) {
	service, dbConn := newTestUserService(t)
	defer dbConn.Close(context.Background())
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx, rw := generateTestEchoContextFromRequest(req)
	ctx.SetParamNames("id")
	ctx.SetParamValues("not-a-uuid")

	err := listForUser(ctx, service)
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

func TestIT_UserController_ListForUser(t *testing.T) {
	service, dbConn := newTestUserService(t)
	defer dbConn.Close(context.Background())
	user := insertTestUser(t, dbConn)

	room1 := insertTestRoom(t, dbConn)
	insertTestRoom(t, dbConn)
	insertUserInRoom(t, dbConn, user.Id, room1.Id)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx, rw := generateTestEchoContextFromRequest(req)
	ctx.SetParamNames("id")
	ctx.SetParamValues(user.Id.String())

	err := listForUser(ctx, service)
	assert.Nil(t, err, "Actual err: %v", err)

	var responseDto []communication.RoomDtoResponse
	err = json.Unmarshal(rw.Body.Bytes(), &responseDto)
	assert.Nil(t, err, "Actual err: %v", err)

	expected := []communication.RoomDtoResponse{
		communication.ToRoomDtoResponse(room1),
	}
	assert.ElementsMatch(t, expected, responseDto)
}

func TestIT_UserController_ListForUser_WhenUserHasNoRoom_ExpectEmptySlice(t *testing.T) {
	service, dbConn := newTestUserService(t)
	defer dbConn.Close(context.Background())
	user := insertTestUser(t, dbConn)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx, rw := generateTestEchoContextFromRequest(req)
	ctx.SetParamNames("id")
	ctx.SetParamValues(user.Id.String())

	err := listForUser(ctx, service)
	assert.Nil(t, err, "Actual err: %v", err)

	var responseDto []communication.RoomDtoResponse
	err = json.Unmarshal(rw.Body.Bytes(), &responseDto)
	assert.Nil(t, err, "Actual err: %v", err)

	assert.Equal(t, []communication.RoomDtoResponse{}, responseDto)
}

func TestIT_UserController_DeleteUser_WhenIdHasWrongSyntax_ExpectBadRequest(t *testing.T) {
	service, dbConn := newTestUserService(t)
	defer dbConn.Close(context.Background())
	req := httptest.NewRequest(http.MethodDelete, "/not-a-uuid", nil)
	ctx, rw := generateTestEchoContextFromRequest(req)

	err := deleteUser(ctx, service)
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

func TestIT_UserController_DeleteUser(t *testing.T) {
	service, dbConn := newTestUserService(t)
	defer dbConn.Close(context.Background())
	user := insertTestUser(t, dbConn)

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	ctx, rw := generateTestEchoContextFromRequest(req)
	ctx.SetParamNames("id")
	ctx.SetParamValues(user.Id.String())

	err := deleteUser(ctx, service)
	assert.Nil(t, err, "Actual err: %v", err)

	assert.Equal(t, http.StatusNoContent, rw.Code)
	assertUserDoesNotExist(t, dbConn, user.Id)
}

func TestIT_UserController_DeleteUser_WhenUserDoesNotExist_ExpectSuccess(t *testing.T) {
	service, dbConn := newTestUserService(t)
	defer dbConn.Close(context.Background())

	nonExistingId := uuid.MustParse("00000000-0000-1221-0000-000000000000")
	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	ctx, rw := generateTestEchoContextFromRequest(req)
	ctx.SetParamNames("id")
	ctx.SetParamValues(nonExistingId.String())

	err := deleteUser(ctx, service)
	assert.Nil(t, err, "Actual err: %v", err)

	assert.Equal(t, http.StatusNoContent, rw.Code)
}

func newTestUserService(t *testing.T) (service.UserService, db.Connection) {
	dbConn := newTestDbConnection(t)
	repos := repositories.New(dbConn)
	return service.NewUserService(dbConn, repos), dbConn
}

func insertTestUser(t *testing.T, conn db.Connection) persistence.User {
	name := fmt.Sprintf("my-user-%s", uuid.NewString())
	return insertTestUserWithName(t, conn, name)
}

func insertTestUserWithName(t *testing.T, conn db.Connection, name string) persistence.User {
	repo := repositories.NewUserRepository(conn)

	tx, err := conn.BeginTx(context.Background())
	assert.Nil(t, err, "Actual err: %v", err)

	id := uuid.New()
	user := persistence.User{
		Id:      id,
		Name:    name,
		ApiUser: uuid.New(),
	}
	out, err := repo.Create(context.Background(), tx, user)
	tx.Close(context.Background())
	assert.Nil(t, err, "Actual err: %v", err)

	assertUserExists(t, conn, out.Id)

	return out
}

func assertUserExists(t *testing.T, dbConn db.Connection, id uuid.UUID) {
	value, err := db.QueryOne[uuid.UUID](
		context.Background(),
		dbConn,
		"SELECT id FROM chat_user WHERE id = $1",
		id,
	)
	assert.Nil(t, err, "Actual err: %v", err)
	assert.Equal(t, id, value)
}

func assertUserDoesNotExist(t *testing.T, conn db.Connection, id uuid.UUID) {
	value, err := db.QueryOne[int](
		context.Background(),
		conn,
		"SELECT COUNT(id) FROM chat_user WHERE id = $1",
		id,
	)
	assert.Nil(t, err, "Actual err: %v", err)
	assert.Zero(t, value)
}
