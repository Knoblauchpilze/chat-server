package repositories

import (
	"context"
	"time"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/db"
	"github.com/Knoblauchpilze/backend-toolkit/pkg/db/pgx"
	"github.com/Knoblauchpilze/backend-toolkit/pkg/errors"
	"github.com/Knoblauchpilze/chat-server/pkg/persistence"
	"github.com/google/uuid"
)

type MessageRepository interface {
	Create(ctx context.Context, msg persistence.Message) (persistence.Message, error)
	ListForRoom(ctx context.Context, room uuid.UUID) ([]persistence.Message, error)
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

const userNotInRoomForeignKey = "message_chat_user_room_fkey"

func (r *messageRepositoryImpl) Create(
	ctx context.Context, msg persistence.Message,
) (persistence.Message, error) {
	createdAt, err := db.QueryOne[time.Time](
		ctx,
		r.conn,
		createMessageSqlTemplate,
		msg.Id,
		msg.ChatUser,
		msg.Room,
		msg.Message,
	)

	msg.CreatedAt = createdAt.UTC()

	if errors.IsErrorWithCode(err, pgx.ForeignKeyValidation) {
		foreignKey, ok := extractForeignKeyViolation(err)

		if ok && foreignKey == userNotInRoomForeignKey {
			return msg, errors.WrapCode(err, ErrUserNotRegisteredInRoom)
		}
	}

	return msg, err
}

const listMessageByRoomSqlTemplate = `
SELECT
	m.id,
	m.chat_user,
	m.room,
	m.message,
	m.created_at
FROM
	message AS m
	LEFT JOIN room AS r ON m.room = r.id
WHERE
	m.room = $1`

func (r *messageRepositoryImpl) ListForRoom(
	ctx context.Context, room uuid.UUID,
) ([]persistence.Message, error) {
	messages, err := db.QueryAll[persistence.Message](
		ctx,
		r.conn,
		listMessageByRoomSqlTemplate,
		room,
	)

	if err == nil {
		for id, message := range messages {
			messages[id].CreatedAt = message.CreatedAt.UTC()
		}
	}

	return messages, err
}
