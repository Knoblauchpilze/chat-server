package repositories

import (
	"context"
	"time"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/db"
	"github.com/Knoblauchpilze/chat-server/pkg/persistence"
)

type MessageRepository interface {
	Create(ctx context.Context, msg persistence.Message) (persistence.Message, error)
}

type messageRepositoryImpl struct {
	conn db.Connection
}

func NewMessageRepository(conn db.Connection) MessageRepository {
	return &messageRepositoryImpl{
		conn: conn,
	}
}

const createMessageSqlTemplate = `
INSERT INTO message (id, chat_user, room, message)
	VALUES ($1, $2, $3, $4)
	RETURNING created_at`

func (r *messageRepositoryImpl) Create(
	ctx context.Context, msg persistence.Message,
) (persistence.Message, error) {
	createdAt, err := db.QueryOne[time.Time](
		ctx,
		r.conn,
		createMessageSqlTemplate,
		msg.Id,
		msg.User,
		msg.Room,
		msg.Message,
	)
	msg.CreatedAt = createdAt
	return msg, err
}
