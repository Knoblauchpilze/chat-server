package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/db"
	"github.com/Knoblauchpilze/backend-toolkit/pkg/db/postgresql"
	"github.com/Knoblauchpilze/chat-server/pkg/persistence"
	"github.com/Knoblauchpilze/chat-server/pkg/repositories"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

const (
	success = "SUCCESS"
	failure = "ERROR"
)

const (
	reasonableWaitTimeForServerToBeUp = 200 * time.Millisecond
	reasonableReadTimeout             = 100 * time.Millisecond
	reasonableReadSizeInBytes         = 1024
)

var dbTestConfig = postgresql.NewConfigForLocalhost(
	"db_chat_server",
	"chat_server_manager",
	"manager_password",
)

var errSample = fmt.Errorf("some error")
var sampleData = []byte("hello\n")

func asyncCancelContext(delay time.Duration, cancel context.CancelFunc) {
	go func() {
		time.Sleep(delay)
		cancel()
	}()
}

func assertConnectionIsClosed(t *testing.T, conn net.Conn) {
	conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))

	oneByte := make([]byte, 1)
	_, err := conn.Read(oneByte)
	assert.Equal(t, io.EOF, err, "Actual err: %v", err)
}

func assertConnectionIsStillOpen(t *testing.T, conn net.Conn) {
	conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))

	oneByte := make([]byte, 1)
	_, err := conn.Read(oneByte)

	opErr, ok := err.(*net.OpError)
	assert.True(t, ok, "Actual err: %v", err)
	if ok {
		assert.True(t, opErr.Timeout(), "Actual err: %v", opErr)
	}
}

func readFromConnection(t *testing.T, conn net.Conn) []byte {
	conn.SetReadDeadline(time.Now().Add(reasonableReadTimeout))

	out := make([]byte, reasonableReadSizeInBytes)
	n, err := conn.Read(out)
	assert.Nil(t, err, "Actual err: %v", err)

	return out[:n]
}

func assertNoDataReceived(t *testing.T, conn net.Conn) {
	conn.SetReadDeadline(time.Now().Add(reasonableReadTimeout))

	oneByte := make([]byte, 1)
	_, err := conn.Read(oneByte)

	assert.True(t, isTimeoutError(err), "Actual err: %v", err)
}

func drainConnection(t *testing.T, conn net.Conn) []byte {
	conn.SetReadDeadline(time.Now().Add(reasonableReadTimeout))

	out := make([]byte, reasonableReadSizeInBytes)
	n, err := conn.Read(out)
	if err != nil && err != io.EOF && !isTimeoutError(err) {
		assert.Nil(t, err, "Actual err: %v", err)
	}

	return out[:n]
}

func isTimeoutError(err error) bool {
	opErr, ok := err.(*net.OpError)
	if !ok {
		return false
	}

	return opErr.Timeout()
}

func newTestDbConnection(t *testing.T) db.Connection {
	conn, err := db.New(context.Background(), dbTestConfig)
	assert.Nil(t, err, "Actual err: %v", err)
	return conn
}

func insertTestUser(t *testing.T, dbConn db.Connection) persistence.User {
	repo := repositories.NewUserRepository(dbConn)

	tx, err := dbConn.BeginTx(context.Background())
	assert.Nil(t, err, "Actual err: %v", err)

	id := uuid.New()
	user := persistence.User{
		Id:        id,
		Name:      fmt.Sprintf("my-user-%s", id),
		ApiUser:   uuid.New(),
		CreatedAt: time.Now(),
	}
	out, err := repo.Create(context.Background(), tx, user)
	tx.Close(context.Background())
	assert.Nil(t, err, "Actual err: %v", err)

	assertUserExists(t, dbConn, id)

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

func insertTestRoom(t *testing.T, dbConn db.Connection) persistence.Room {
	repo := repositories.NewRoomRepository(dbConn)

	id := uuid.New()
	room := persistence.Room{
		Id:        id,
		Name:      fmt.Sprintf("my-room-%s", id),
		CreatedAt: time.Now(),
	}
	out, err := repo.Create(context.Background(), room)
	assert.Nil(t, err, "Actual err: %v", err)

	assertRoomExists(t, dbConn, id)

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

func doRequest(
	t *testing.T, method string, url string,
) *http.Response {
	return doRequestWithBody(t, method, url, nil)
}

func doRequestWithData[T any](
	t *testing.T, method string, url string, data T,
) *http.Response {
	var body bytes.Buffer
	err := json.NewEncoder(&body).Encode(data)
	assert.Nil(t, err, "Actual err: %v", err)

	return doRequestWithBody(t, method, url, body.Bytes())
}

func doRequestWithBody(
	t *testing.T, method string, url string, body []byte,
) *http.Response {
	client := &http.Client{}

	req, err := http.NewRequest(method, url, bytes.NewReader(body))
	assert.Nil(t, err, "Actual err: %v", err)

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	rw, err := client.Do(req)
	assert.Nil(t, err, "Actual err: %v", err)

	return rw
}

func assertResponseAndExtractDetails[T any](
	t *testing.T, rw *http.Response, status string,
) T {
	type response struct {
		Status  string          `json:"status"`
		Details json.RawMessage `json:"details"`
	}

	data, err := io.ReadAll(rw.Body)
	assert.Nil(t, err, "Actual err: %v", err)

	var responseData response
	err = json.Unmarshal(data, &responseData)
	assert.Nil(t, err, "Actual err: %v", err)

	assert.Equal(t, status, responseData.Status)

	var out T
	err = json.Unmarshal(responseData.Details, &out)
	assert.Nil(t, err, "Actual err: %v", err)

	return out
}

func connectToServerAndSendHandshake(
	t *testing.T, port uint16, dbConn db.Connection,
) (net.Conn, persistence.User) {
	user := insertTestUser(t, dbConn)

	conn, err := net.Dial("tcp", fmt.Sprintf(":%d", port))
	assert.Nil(t, err, "Actual err: %v", err)

	n, err := conn.Write(user.Id[:])
	assert.Nil(t, err, "Actual err: %v", err)
	assert.Equal(t, len(user.Id), n)

	return conn, user
}
