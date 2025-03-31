package clients

import (
	"context"
	"testing"
	"time"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/db"
	"github.com/Knoblauchpilze/backend-toolkit/pkg/errors"
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

func TestIT_Handshake_WhenDataIsNotAnId_ExpectFailure(t *testing.T) {
	handshake, dbConn := newTestHandshake(t)
	defer dbConn.Close(context.Background())
	client, server := newTestConnection(t, 7402)

	expected := uuid.New()
	n, err := client.Write(expected[:])
	assert.Equal(t, len(expected), n)
	assert.Nil(t, err, "Actual err: %v", err)

	actual, err := handshake.Perform(server)

	assert.Nil(t, err, "Actual err: %v", err)
	assert.Equal(t, expected, actual)
}

func newTestHandshake(t *testing.T) (Handshake, db.Connection) {
	dbConn := newTestDbConnection(t)
	userRepo := repositories.NewUserRepository(dbConn)

	handshake := NewHandshake(
		userRepo, reasonableHandshakeTimeout,
	)

	return handshake, dbConn
}
