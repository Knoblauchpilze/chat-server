package service

import (
	"context"
	"fmt"
	"testing"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/db"
	"github.com/Knoblauchpilze/backend-toolkit/pkg/db/postgresql"
	"github.com/Knoblauchpilze/chat-server/pkg/persistence"
	"github.com/Knoblauchpilze/chat-server/pkg/repositories"
	"github.com/google/uuid"
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

func insertTestMessage(
	t *testing.T,
	conn db.Connection,
	user uuid.UUID,
	room uuid.UUID,
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
