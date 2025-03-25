package repositories

import (
	"context"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/db"
	"github.com/Knoblauchpilze/chat-server/pkg/persistence"
	"github.com/google/uuid"
)

type RoomRepository interface {
	Create(ctx context.Context, room persistence.Room) (persistence.Room, error)
	Get(ctx context.Context, id uuid.UUID) (persistence.Room, error)
	ListForUser(ctx context.Context, user uuid.UUID) ([]persistence.Room, error)
	Delete(ctx context.Context, tx db.Transaction, id uuid.UUID) error
}

type roomRepositoryImpl struct {
	conn db.Connection
}

func NewRoomRepository(conn db.Connection) RoomRepository {
	return &roomRepositoryImpl{
		conn: conn,
	}
}

const createRoomSqlTemplate = `
INSERT INTO room (id, name)
	VALUES ($1, $2)
	RETURNING created_at, updated_at`

func (r *roomRepositoryImpl) Create(
	ctx context.Context, room persistence.Room,
) (persistence.Room, error) {
	times, err := db.QueryOne[createdAtUpdatedAt](
		ctx,
		r.conn,
		createRoomSqlTemplate,
		room.Id,
		room.Name,
	)
	room.CreatedAt = times.CreatedAt
	room.UpdatedAt = times.UpdatedAt
	return room, err
}

const getRoomSqlTemplate = `
SELECT
	id,
	name,
	created_at,
	updated_at
FROM
	room
WHERE
	id = $1`

func (r *roomRepositoryImpl) Get(
	ctx context.Context, id uuid.UUID,
) (persistence.Room, error) {
	return db.QueryOne[persistence.Room](ctx, r.conn, getRoomSqlTemplate, id)
}

const listForUserSqlTemplate = `
SELECT
	r.id,
	r.name,
	r.created_at,
	r.updated_at
FROM
	room AS r
	LEFT JOIN room_user AS ru ON r.id = ru.room
WHERE
	ru.chat_user = $1`

func (r *roomRepositoryImpl) ListForUser(
	ctx context.Context, user uuid.UUID,
) ([]persistence.Room, error) {
	return db.QueryAll[persistence.Room](ctx, r.conn, listForUserSqlTemplate, user)
}

const deleteRoomSqlTemplate = `
DELETE FROM
	room
WHERE
	id = $1`

func (r *roomRepositoryImpl) Delete(
	ctx context.Context, tx db.Transaction, id uuid.UUID,
) error {
	_, err := tx.Exec(ctx, deleteRoomSqlTemplate, id)
	return err
}
