package repositories

import (
	"context"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/db"
	"github.com/Knoblauchpilze/chat-server/pkg/persistence"
	"github.com/google/uuid"
)

type UserRepository interface {
	Create(ctx context.Context, user persistence.User) (persistence.User, error)
	Get(ctx context.Context, id uuid.UUID) (persistence.User, error)
	ListForRoom(ctx context.Context, room uuid.UUID) ([]persistence.User, error)
	Delete(ctx context.Context, tx db.Transaction, id uuid.UUID) error
}

type userRepositoryImpl struct {
	conn db.Connection
}

func NewUserRepository(conn db.Connection) UserRepository {
	return &userRepositoryImpl{
		conn: conn,
	}
}

const createUserSqlTemplate = `
INSERT INTO chat_user (id, name, api_user)
	VALUES ($1, $2, $3)
	RETURNING created_at, updated_at`

func (r *userRepositoryImpl) Create(
	ctx context.Context, user persistence.User,
) (persistence.User, error) {
	times, err := db.QueryOne[createdAtUpdatedAt](
		ctx,
		r.conn,
		createUserSqlTemplate,
		user.Id,
		user.Name,
		user.ApiUser,
	)
	user.CreatedAt = times.CreatedAt
	user.UpdatedAt = times.UpdatedAt
	return user, err
}

const getUserSqlTemplate = `
SELECT
	id,
	name,
	api_user,
	created_at,
	updated_at,
	version
FROM
	chat_user
WHERE
	id = $1`

func (r *userRepositoryImpl) Get(
	ctx context.Context, id uuid.UUID,
) (persistence.User, error) {
	return db.QueryOne[persistence.User](ctx, r.conn, getUserSqlTemplate, id)
}

const listUserByRoomSqlTemplate = `
SELECT
	cu.id,
	cu.name,
	cu.api_user,
	cu.created_at,
	cu.updated_at,
	cu.version
FROM
	room_user AS ru
	LEFT JOIN chat_user AS cu on ru.chat_user = cu.id
WHERE
	ru.room = $1`

func (r *userRepositoryImpl) ListForRoom(
	ctx context.Context, room uuid.UUID,
) ([]persistence.User, error) {
	return db.QueryAll[persistence.User](
		ctx,
		r.conn,
		listUserByRoomSqlTemplate,
		room,
	)
}

const deleteUserSqlTemplate = `
DELETE FROM
	chat_user
WHERE
	id = $1`

func (r *userRepositoryImpl) Delete(
	ctx context.Context, tx db.Transaction, id uuid.UUID,
) error {
	_, err := tx.Exec(ctx, deleteUserSqlTemplate, id)
	return err
}
