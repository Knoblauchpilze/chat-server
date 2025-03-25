package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/db"
	"github.com/Knoblauchpilze/backend-toolkit/pkg/db/pgx"
	"github.com/Knoblauchpilze/backend-toolkit/pkg/errors"
	"github.com/Knoblauchpilze/chat-server/pkg/persistence"
	eassert "github.com/Knoblauchpilze/easy-assert/assert"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestIT_UserRepository_Create(t *testing.T) {
	repo, conn := newTestUserRepository(t)
	startTime := time.Now()

	user := persistence.User{
		Id:      uuid.New(),
		Name:    "my-name-" + uuid.New().String(),
		ApiUser: uuid.New(),
	}

	actual, err := repo.Create(context.Background(), user)
	assert.Nil(t, err, "Actual err: %v", err)

	assert.True(t, eassert.EqualsIgnoringFields(actual, user, "CreatedAt", "UpdatedAt"))
	assert.Equal(t, actual.CreatedAt, actual.UpdatedAt)
	assert.True(t, actual.CreatedAt.After(startTime))
	assertUserExists(t, conn, user.Id)
}

func TestIT_UserRepository_Create_WhenDuplicateName_ExpectFailure(t *testing.T) {
	repo, conn := newTestUserRepository(t)
	user := insertTestUser(t, conn)

	newUser := persistence.User{
		Id:        uuid.New(),
		Name:      user.Name,
		ApiUser:   uuid.New(),
		CreatedAt: time.Now(),
	}

	_, err := repo.Create(context.Background(), newUser)

	assert.True(
		t,
		errors.IsErrorWithCode(err, pgx.UniqueConstraintViolation),
		"Actual err: %v",
		err,
	)
	assertUserDoesNotExist(t, conn, newUser.Id)
}

func TestIT_UserRepository_Get(t *testing.T) {
	repo, conn := newTestUserRepository(t)
	user := insertTestUser(t, conn)

	actual, err := repo.Get(context.Background(), user.Id)
	assert.Nil(t, err, "Actual err: %v", err)

	assert.True(t, eassert.EqualsIgnoringFields(actual, user))
}

func TestIT_UserRepository_Get_WhenNotFound_ExpectFailure(t *testing.T) {
	repo, _ := newTestUserRepository(t)

	// Non-existent id
	id := uuid.MustParse("00000000-1111-2222-1111-000000000000")
	_, err := repo.Get(context.Background(), id)
	assert.True(
		t,
		errors.IsErrorWithCode(err, db.NoMatchingRows),
		"Actual err: %v",
		err,
	)
}

func TestIT_UserRepository_ListForRoom(t *testing.T) {
	repo, conn := newTestUserRepository(t)
	room := insertTestRoom(t, conn)
	user1 := insertTestUser(t, conn)
	insertTestUser(t, conn)
	registerUserInRoom(t, conn, user1.Id, room.Id)

	actual, err := repo.ListForRoom(context.Background(), room.Id)
	assert.Nil(t, err, "Actual err: %v", err)

	assert.Len(t, actual, 1)
	assert.True(t, eassert.EqualsIgnoringFields(actual[0], user1))
}

func TestIT_UserRepository_ListForRoom_WhenNoUserRegistered_ReturnsEmptySlice(t *testing.T) {
	repo, conn := newTestUserRepository(t)
	room := insertTestRoom(t, conn)

	actual, err := repo.ListForRoom(context.Background(), room.Id)
	assert.Nil(t, err, "Actual err: %v", err)

	assert.Equal(t, []persistence.User{}, actual)
}

func TestIT_UserRepository_Delete(t *testing.T) {
	repo, conn, tx := newTestUserRepositoryAndTransaction(t)

	user := insertTestUser(t, conn)

	err := repo.Delete(context.Background(), tx, user.Id)
	tx.Close(context.Background())

	assert.Nil(t, err, "Actual err: %v", err)
	assertUserDoesNotExist(t, conn, user.Id)
}

func TestIT_UserRepository_Delete_WhenNotFound_ExpectSuccess(t *testing.T) {
	repo, conn, tx := newTestUserRepositoryAndTransaction(t)

	user := insertTestUser(t, conn)
	id := uuid.New()
	assert.NotEqual(t, user.Id, id)

	err := repo.Delete(context.Background(), tx, id)
	tx.Close(context.Background())

	assert.Nil(t, err, "Actual err: %v", err)
	assertUserExists(t, conn, user.Id)
}

func newTestUserRepository(t *testing.T) (UserRepository, db.Connection) {
	conn := newTestConnection(t)
	return NewUserRepository(conn), conn
}

func newTestUserRepositoryAndTransaction(t *testing.T) (UserRepository, db.Connection, db.Transaction) {
	conn := newTestConnection(t)
	tx, err := conn.BeginTx(context.Background())
	assert.Nil(t, err, "Actual err: %v", err)
	return NewUserRepository(conn), conn, tx
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

func assertUserDoesNotExist(t *testing.T, conn db.Connection, id uuid.UUID) {
	value, err := db.QueryOne[int](
		context.Background(),
		conn,
		"SELECT COUNT(id) FROM chat_user WHERE id = $1",
		id,
	)
	assert.Nil(t, err, "Actual err: %v", err)
	assert.Zero(t, value)
}

func insertTestUser(t *testing.T, conn db.Connection) persistence.User {
	someTime := time.Date(2025, 3, 2, 10, 17, 15, 0, time.UTC)

	user := persistence.User{
		Id:        uuid.New(),
		Name:      "my-name-" + uuid.New().String(),
		ApiUser:   uuid.New(),
		CreatedAt: someTime,
	}

	updatedAt, err := db.QueryOne[time.Time](
		context.Background(),
		conn,
		`INSERT INTO
			chat_user (id, name, api_user, created_at)
			VALUES ($1, $2, $3, $4)
			RETURNING updated_at`,
		user.Id,
		user.Name,
		user.ApiUser,
		user.CreatedAt,
	)
	assert.Nil(t, err, "Actual err: %v", err)

	user.UpdatedAt = updatedAt

	return user
}
