package repositories

import (
	"context"
	"time"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/db"
	"github.com/Knoblauchpilze/backend-toolkit/pkg/errors"
	"github.com/Knoblauchpilze/chat-server/pkg/persistence"
	"github.com/google/uuid"
)

type RoomRepository interface {
	Create(ctx context.Context, room persistence.Room) (persistence.Room, error)
	Get(ctx context.Context, id uuid.UUID) (persistence.Room, error)
	ListForUser(ctx context.Context, user uuid.UUID) ([]persistence.Room, error)
	RegisterUserInRoom(ctx context.Context, tx db.Transaction, user uuid.UUID, room uuid.UUID) error
	RegisterUserInRoomByName(ctx context.Context, tx db.Transaction, user uuid.UUID, room string) error
	Delete(ctx context.Context, tx db.Transaction, id uuid.UUID) error
	DeleteUserFromRoomByName(ctx context.Context, tx db.Transaction, user uuid.UUID, room string) error
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

	// https://www.reddit.com/r/golang/comments/1gbvowf/dealing_with_timezone_issues_when_running_unit/
	room.CreatedAt = times.CreatedAt.UTC()
	room.UpdatedAt = times.UpdatedAt.UTC()

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
	room, err := db.QueryOne[persistence.Room](ctx, r.conn, getRoomSqlTemplate, id)

	if err == nil {
		room.CreatedAt = room.CreatedAt.UTC()
		room.UpdatedAt = room.UpdatedAt.UTC()
	}

	return room, err
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
	rooms, err := db.QueryAll[persistence.Room](ctx, r.conn, listForUserSqlTemplate, user)

	if err == nil {
		for id, room := range rooms {
			rooms[id].CreatedAt = room.CreatedAt.UTC()
			rooms[id].UpdatedAt = room.UpdatedAt.UTC()
		}
	}

	return rooms, err
}

const registerUserInRoomSqlTemplate = `
INSERT INTO room_user (chat_user, room)
	VALUES ($1, $2)
	RETURNING created_at`

func (r *roomRepositoryImpl) RegisterUserInRoom(
	ctx context.Context, tx db.Transaction, user uuid.UUID, room uuid.UUID,
) error {
	_, err := tx.Exec(ctx, registerUserInRoomSqlTemplate, user, room)
	return err
}

const registerUserInRoomByNameSqlTemplate = `
INSERT INTO
	room_user (chat_user, room)
SELECT
	$1,
	id
FROM
	room
WHERE
	name = $2
RETURNING created_at`

func (r *roomRepositoryImpl) RegisterUserInRoomByName(
	ctx context.Context, tx db.Transaction, user uuid.UUID, room string,
) error {
	_, err := db.QueryOneTx[time.Time](
		ctx, tx, registerUserInRoomByNameSqlTemplate, user, room,
	)

	if errors.IsErrorWithCode(err, db.NoMatchingRows) {
		return errors.NewCode(ErrNoSuchRoom)
	}

	return err
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

const deleteUserFromByNameSqlTemplate = `
DELETE FROM
	room_user AS rud
USING
	room_user AS ru
	LEFT JOIN room AS r ON r.id = ru.room
WHERE
	rud.room = ru.room
	AND rud.chat_user = ru.chat_user
	AND ru.chat_user = $1
	AND r.name = $2`

func (r *roomRepositoryImpl) DeleteUserFromRoomByName(
	ctx context.Context, tx db.Transaction, user uuid.UUID, room string,
) error {
	_, err := tx.Exec(ctx, deleteUserFromByNameSqlTemplate, user, room)
	return err
}
