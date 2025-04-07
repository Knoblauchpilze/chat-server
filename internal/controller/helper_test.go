package controller

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/db"
	"github.com/Knoblauchpilze/backend-toolkit/pkg/db/postgresql"
	"github.com/Knoblauchpilze/chat-server/pkg/persistence"
	"github.com/Knoblauchpilze/chat-server/pkg/repositories"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

var dbTestConfig = postgresql.NewConfigForLocalhost(
	"db_chat_server",
	"chat_server_manager",
	"manager_password",
)

func newTestDbConnection(t *testing.T) db.Connection {
	conn, err := db.New(context.Background(), dbTestConfig)
	assert.Nil(t, err, "Actual err: %v", err)
	return conn
}

func generateTestEchoContextFromRequest(req *http.Request) (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	rw := httptest.NewRecorder()

	ctx := e.NewContext(req, rw)
	return ctx, rw
}

func generateTestRequestWithQueryParam(
	method string, key string, param string,
) *http.Request {
	req := httptest.NewRequest(method, "/", nil)
	query := req.URL.Query()
	query.Add(key, param)
	req.URL.RawQuery = query.Encode()
	return req
}

func insertUserInRoom(t *testing.T, conn db.Connection, user uuid.UUID, room uuid.UUID) {
	sqlQuery := `INSERT INTO room_user (room, chat_user) VALUES ($1, $2)`

	count, err := conn.Exec(
		context.Background(),
		sqlQuery,
		room,
		user,
	)
	assert.Nil(t, err, "Actual err: %v", err)
	assert.Equal(t, int64(1), count)
}

func insertTestMessage(
	t *testing.T, conn db.Connection, user uuid.UUID, room uuid.UUID,
) persistence.Message {
	repo := repositories.NewMessageRepository(conn)

	id := uuid.New()
	msg := persistence.Message{
		Id:       id,
		ChatUser: user,
		Room:     room,
		Message:  fmt.Sprintf("my-message-%s", id),
	}
	out, err := repo.Create(context.Background(), msg)
	assert.Nil(t, err, "Actual err: %v", err)

	assertMessageExists(t, conn, out.Id)

	return out
}

func assertMessageExists(t *testing.T, conn db.Connection, id uuid.UUID) {
	value, err := db.QueryOne[uuid.UUID](
		context.Background(),
		conn,
		"SELECT id FROM message WHERE id = $1",
		id,
	)
	assert.Nil(t, err, "Actual err: %v", err)
	assert.Equal(t, id, value)
}
