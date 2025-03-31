package clients

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/db"
	"github.com/Knoblauchpilze/backend-toolkit/pkg/errors"
	"github.com/Knoblauchpilze/chat-server/pkg/persistence"
	"github.com/Knoblauchpilze/chat-server/pkg/repositories"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

const reasonableHandshakeTimeout = 50 * time.Millisecond

func TestIT_Handshake_WhenNoDataSent_ExpectFailure(t *testing.T) {
	handshake, dbConn := newTestHandshake(t)
	defer dbConn.Close(context.Background())
	_, server := newTestConnection(t, 7400)

	_, err := handshake.Perform(server)

	assert.True(
		t,
		errors.IsErrorWithCode(err, HandshakeTimeout),
		"Actual err: %v",
		err,
	)
}

func TestIT_Handshake_WhenPartialDataSent_ExpectFailure(t *testing.T) {
	handshake, dbConn := newTestHandshake(t)
	defer dbConn.Close(context.Background())
	client, server := newTestConnection(t, 7401)

	id := uuid.New()
	n, err := client.Write(id[:5])
	assert.Equal(t, 5, n)
	assert.Nil(t, err, "Actual err: %v", err)

	_, err = handshake.Perform(server)

	assert.True(
		t,
		errors.IsErrorWithCode(err, IncompleteHandshake),
		"Actual err: %v",
		err,
	)
}

func TestIT_Handshake_WhenUserIsNotRegisteredInDb_ExpectFailure(t *testing.T) {
	handshake, dbConn := newTestHandshake(t)
	defer dbConn.Close(context.Background())
	client, server := newTestConnection(t, 7402)

	id := uuid.New()
	n, err := client.Write(id[:])
	assert.Equal(t, len(id), n)
	assert.Nil(t, err, "Actual err: %v", err)

	_, err = handshake.Perform(server)

	assert.True(
		t,
		errors.IsErrorWithCode(err, HandshakeFailure),
		"Actual err: %v",
		err,
	)
	dbErr := errors.Unwrap(err)
	assert.True(
		t,
		errors.IsErrorWithCode(dbErr, db.NoMatchingRows),
		"Actual err: %v",
		dbErr,
	)
}

func TestIT_Handshake_WhenUserIsRegistered_ExpectSuccess(t *testing.T) {
	handshake, dbConn := newTestHandshake(t)
	defer dbConn.Close(context.Background())
	client, server := newTestConnection(t, 7402)

	user := insertTestUser(t, dbConn)

	n, err := client.Write(user.Id[:])
	assert.Equal(t, len(user.Id), n)
	assert.Nil(t, err, "Actual err: %v", err)

	actual, err := handshake.Perform(server)

	assert.Nil(t, err, "Actual err: %v", err)
	assert.Equal(t, user.Id, actual)
}

func newTestHandshake(t *testing.T) (Handshake, db.Connection) {
	dbConn := newTestDbConnection(t)
	userRepo := repositories.NewUserRepository(dbConn)

	handshake := NewHandshake(
		userRepo, reasonableHandshakeTimeout,
	)

	return handshake, dbConn
}

func insertTestUser(t *testing.T, conn db.Connection) persistence.User {
	repo := repositories.NewUserRepository(conn)

	id := uuid.New()
	user := persistence.User{
		Id:      id,
		Name:    fmt.Sprintf("my-user-%s", id),
		ApiUser: uuid.New(),
	}
	out, err := repo.Create(context.Background(), user)
	assert.Nil(t, err, "Actual err: %v", err)

	assertUserExists(t, conn, out.Id)

	return out
}

func assertUserExists(t *testing.T, conn db.Connection, id uuid.UUID) {
	value, err := db.QueryOne[uuid.UUID](
		context.Background(),
		conn,
		"SELECT id FROM chat_user WHERE id = $1",
		id,
	)
	assert.Nil(t, err, "Actual err: %v", err)
	assert.Equal(t, id, value)
}
