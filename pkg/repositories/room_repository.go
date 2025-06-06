package repositories

import (
	"context"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/db"
	"github.com/Knoblauchpilze/backend-toolkit/pkg/db/pgx"
	"github.com/Knoblauchpilze/backend-toolkit/pkg/errors"
	"github.com/Knoblauchpilze/chat-server/pkg/persistence"
	"github.com/google/uuid"
)

type RoomRepository interface {
	Create(ctx context.Context, tx db.Transaction, room persistence.Room) (persistence.Room, error)
	Get(ctx context.Context, id uuid.UUID) (persistence.Room, error)
	List(ctx context.Context) ([]persistence.Room, error)
	UserInRoom(ctx context.Context, user uuid.UUID, room uuid.UUID) (bool, error)
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
	ctx context.Context, tx db.Transaction, room persistence.Room,
) (persistence.Room, error) {
	times, err := db.QueryOneTx[createdAtUpdatedAt](
		ctx,
		tx,
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

const listRoomSqlTemplate = `
SELECT
	id,
	name,
	created_at,
	updated_at
FROM
	room`

func (r *roomRepositoryImpl) List(
	ctx context.Context,
) ([]persistence.Room, error) {
	rooms, err := db.QueryAll[persistence.Room](ctx, r.conn, listRoomSqlTemplate)

	for id := range rooms {
		rooms[id].CreatedAt = rooms[id].CreatedAt.UTC()
		rooms[id].UpdatedAt = rooms[id].UpdatedAt.UTC()
	}

	return rooms, err
}

const userInRoomSqlTemplate = `
SELECT
	COUNT(*)
FROM
	room_user
WHERE
	chat_user = $1
	AND room = $2`

func (r *roomRepositoryImpl) UserInRoom(
	ctx context.Context, user uuid.UUID, room uuid.UUID,
) (bool, error) {

	count, err := db.QueryOne[int](ctx, r.conn, userInRoomSqlTemplate, user, room)
	return count > 0, err
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

const noSuchUserForeignKey = "room_user_chat_user_fkey"
const noSuchRoomForeignKey = "room_user_room_fkey"

func handleRegistrationError(err error) error {

	if errors.IsErrorWithCode(err, pgx.ForeignKeyValidation) {
		if foreignKey, ok := extractForeignKeyViolation(err); ok {
			switch foreignKey {
			case noSuchUserForeignKey:
				return errors.WrapCode(err, ErrNoSuchUser)
			case noSuchRoomForeignKey:
				return errors.WrapCode(err, ErrNoSuchRoom)
			default:
			}
		}
	}

	if errors.IsErrorWithCode(err, pgx.UniqueConstraintViolation) {
		return errors.WrapCode(err, ErrUserAlreadyRegisteredInRoom)
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
