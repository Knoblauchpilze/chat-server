package repositories

import (
	"context"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/db"
	"github.com/Knoblauchpilze/backend-toolkit/pkg/errors"
	"github.com/google/uuid"
)

type RegistrationRepository interface {
	RegisterInRoom(ctx context.Context, tx db.Transaction, user uuid.UUID, room uuid.UUID) error
	RegisterInRoomByName(ctx context.Context, tx db.Transaction, user uuid.UUID, room string) error
	RegisterByNameInRoom(ctx context.Context, tx db.Transaction, user string, room uuid.UUID) error
	DeleteForRoom(ctx context.Context, tx db.Transaction, room uuid.UUID) error
	DeleteFromRoom(ctx context.Context, tx db.Transaction, room uuid.UUID, user uuid.UUID) error
}

type registrationRepositoryImpl struct{}

func NewRegistrationRepository() RegistrationRepository {
	return &registrationRepositoryImpl{}
}

const registerInRoomSqlTemplate = `
INSERT INTO room_user (chat_user, room)
	VALUES ($1, $2)
	RETURNING created_at`

func (r *registrationRepositoryImpl) RegisterInRoom(
	ctx context.Context, tx db.Transaction, user uuid.UUID, room uuid.UUID,
) error {
	_, err := tx.Exec(ctx, registerInRoomSqlTemplate, user, room)
	return handleRegistrationError(err)
}

const registerInRoomByNameSqlTemplate = `
INSERT INTO
	room_user (chat_user, room)
SELECT
	$1,
	id
FROM
	room
WHERE
	name = $2`

func (r *registrationRepositoryImpl) RegisterInRoomByName(
	ctx context.Context, tx db.Transaction, user uuid.UUID, room string,
) error {
	inserted, err := tx.Exec(ctx, registerInRoomByNameSqlTemplate, user, room)

	if err == nil && inserted == 0 {
		return errors.WrapCode(err, ErrNoSuchRoom)
	}
	return handleRegistrationError(err)
}

const registerByNameInRoomSqlTemplate = `
INSERT INTO
	room_user (chat_user, room)
SELECT
	id,
	$2
FROM
	chat_user
WHERE
	name = $1`

func (r *registrationRepositoryImpl) RegisterByNameInRoom(
	ctx context.Context, tx db.Transaction, user string, room uuid.UUID,
) error {
	inserted, err := tx.Exec(ctx, registerByNameInRoomSqlTemplate, user, room)

	if err == nil && inserted == 0 {
		return errors.WrapCode(err, ErrNoSuchUser)
	}
	return handleRegistrationError(err)
}

const deleteForRoomSqlTemplate = `
DELETE FROM
	room_user
WHERE
	room = $1`

func (r *registrationRepositoryImpl) DeleteForRoom(
	ctx context.Context, tx db.Transaction, room uuid.UUID,
) error {
	_, err := tx.Exec(ctx, deleteForRoomSqlTemplate, room)
	return err
}

const deleteFromRoomSqlTemplate = `
DELETE FROM
	room_user
WHERE
	room = $1
	AND chat_user = $2`

func (r *registrationRepositoryImpl) DeleteFromRoom(
	ctx context.Context, tx db.Transaction, room uuid.UUID, user uuid.UUID,
) error {
	_, err := tx.Exec(ctx, deleteFromRoomSqlTemplate, room, user)
	return err
}
