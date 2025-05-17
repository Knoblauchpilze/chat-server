package repositories

import (
	"context"
	"fmt"
	"testing"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/db"
	"github.com/Knoblauchpilze/backend-toolkit/pkg/errors"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestIT_RegistrationRepository_RegisterUserInRoom(t *testing.T) {
	repo, conn, tx := newTestRegistrationRepositoryAndTransaction(t)
	defer conn.Close(context.Background())
	user := insertTestUser(t, conn)
	room := insertTestRoom(t, conn)

	err := repo.RegisterInRoom(context.Background(), tx, user.Id, room.Id)
	tx.Close(context.Background())
	assert.Nil(t, err, "Actual err: %v", err)

	assertUserRegisteredInRoom(t, conn, user.Id, room.Id)
}

func TestIT_RegistrationRepository_RegisterUserInRoom_WhenRoomDoesNotExist_ExpectFailure(t *testing.T) {
	repo, conn, tx := newTestRegistrationRepositoryAndTransaction(t)
	defer conn.Close(context.Background())
	user := insertTestUser(t, conn)
	room := uuid.New()

	err := repo.RegisterInRoom(context.Background(), tx, user.Id, room)
	tx.Close(context.Background())
	assert.True(
		t,
		errors.IsErrorWithCode(err, ErrNoSuchRoom),
		"Actual err: %v",
		err,
	)

	assertUserNotRegisteredInRoom(t, conn, user.Id, room)
}

func TestIT_RegistrationRepository_RegisterUserInRoom_WhenUserDoesNotExist_ExpectFailure(t *testing.T) {
	repo, conn, tx := newTestRegistrationRepositoryAndTransaction(t)
	defer conn.Close(context.Background())
	user := uuid.New()
	room := insertTestRoom(t, conn)

	err := repo.RegisterInRoom(context.Background(), tx, user, room.Id)
	tx.Close(context.Background())
	assert.True(
		t,
		errors.IsErrorWithCode(err, ErrNoSuchUser),
		"Actual err: %v",
		err,
	)

	assertUserNotRegisteredInRoom(t, conn, user, room.Id)
}

func TestIT_RegistrationRepository_RegisterUserInRoom_WhenUserAlreadyRegistered_ExpectFailure(t *testing.T) {
	repo, conn, tx := newTestRegistrationRepositoryAndTransaction(t)
	defer conn.Close(context.Background())
	user := insertTestUser(t, conn)
	room := insertTestRoom(t, conn)
	registerUserInRoom(t, conn, user.Id, room.Id)

	err := repo.RegisterInRoom(context.Background(), tx, user.Id, room.Id)
	tx.Close(context.Background())
	assert.True(
		t,
		errors.IsErrorWithCode(err, ErrUserAlreadyRegisteredInRoom),
		"Actual err: %v",
		err,
	)
}

func TestIT_RegistrationRepository_RegisterUserInRoomByName(t *testing.T) {
	repo, conn, tx := newTestRegistrationRepositoryAndTransaction(t)
	defer conn.Close(context.Background())
	user := insertTestUser(t, conn)
	room := insertTestRoom(t, conn)

	err := repo.RegisterInRoomByName(context.Background(), tx, user.Id, room.Name)
	tx.Close(context.Background())
	assert.Nil(t, err, "Actual err: %v", err)

	assertUserRegisteredInRoom(t, conn, user.Id, room.Id)
}

func TestIT_RegistrationRepository_RegisterUserInRoomByName_WhenRoomDoesNotExist_ExpectFailure(t *testing.T) {
	repo, conn, tx := newTestRegistrationRepositoryAndTransaction(t)
	defer conn.Close(context.Background())
	user := insertTestUser(t, conn)
	room := fmt.Sprintf("my-inexistent-room-%s", uuid.New().String())

	err := repo.RegisterInRoomByName(context.Background(), tx, user.Id, room)
	tx.Close(context.Background())
	assert.True(
		t,
		errors.IsErrorWithCode(err, ErrNoSuchRoom),
		"Actual err: %v",
		err,
	)
}

func TestIT_RegistrationRepository_RegisterUserInRoomByName_WhenUserDoesNotExist_ExpectFailure(t *testing.T) {
	repo, conn, tx := newTestRegistrationRepositoryAndTransaction(t)
	defer conn.Close(context.Background())
	user := uuid.New()
	room := insertTestRoom(t, conn)

	err := repo.RegisterInRoomByName(context.Background(), tx, user, room.Name)
	tx.Close(context.Background())
	assert.True(
		t,
		errors.IsErrorWithCode(err, ErrNoSuchUser),
		"Actual err: %v",
		err,
	)

	assertUserNotRegisteredInRoom(t, conn, user, room.Id)
}

func TestIT_RegistrationRepository_RegisterUserInRoomByName_WhenUserAlreadyRegistered_ExpectFailure(t *testing.T) {
	repo, conn, tx := newTestRegistrationRepositoryAndTransaction(t)
	defer conn.Close(context.Background())
	user := insertTestUser(t, conn)
	room := insertTestRoom(t, conn)
	registerUserInRoom(t, conn, user.Id, room.Id)

	err := repo.RegisterInRoomByName(context.Background(), tx, user.Id, room.Name)
	tx.Close(context.Background())
	assert.True(
		t,
		errors.IsErrorWithCode(err, ErrUserAlreadyRegisteredInRoom),
		"Actual err: %v",
		err,
	)
}

func TestIT_RegistrationRepository_RegisterUserByNameInRoom(t *testing.T) {
	repo, conn, tx := newTestRegistrationRepositoryAndTransaction(t)
	defer conn.Close(context.Background())
	user := insertTestUser(t, conn)
	room := insertTestRoom(t, conn)

	err := repo.RegisterByNameInRoom(context.Background(), tx, user.Name, room.Id)
	tx.Close(context.Background())
	assert.Nil(t, err, "Actual err: %v", err)

	assertUserRegisteredInRoom(t, conn, user.Id, room.Id)
}

func TestIT_RegistrationRepository_RegisterUserByNameInRoom_WhenRoomDoesNotExist_ExpectFailure(t *testing.T) {
	repo, conn, tx := newTestRegistrationRepositoryAndTransaction(t)
	defer conn.Close(context.Background())
	user := insertTestUser(t, conn)
	room := uuid.New()

	err := repo.RegisterByNameInRoom(context.Background(), tx, user.Name, room)
	tx.Close(context.Background())
	assert.True(
		t,
		errors.IsErrorWithCode(err, ErrNoSuchRoom),
		"Actual err: %v",
		err,
	)
	assertUserNotRegisteredInRoom(t, conn, user.Id, room)
}

func TestIT_RegistrationRepository_RegisterUserByNameInRoom_WhenUserDoesNotExist_ExpectFailure(t *testing.T) {
	repo, conn, tx := newTestRegistrationRepositoryAndTransaction(t)
	defer conn.Close(context.Background())
	user := fmt.Sprintf("my-inexistent-user-%s", uuid.New().String())
	room := insertTestRoom(t, conn)

	err := repo.RegisterByNameInRoom(context.Background(), tx, user, room.Id)
	tx.Close(context.Background())
	assert.True(
		t,
		errors.IsErrorWithCode(err, ErrNoSuchUser),
		"Actual err: %v",
		err,
	)
}

func TestIT_RegistrationRepository_RegisterByNameUserInRoom_WhenUserAlreadyRegistered_ExpectFailure(t *testing.T) {
	repo, conn, tx := newTestRegistrationRepositoryAndTransaction(t)
	defer conn.Close(context.Background())
	user := insertTestUser(t, conn)
	room := insertTestRoom(t, conn)
	registerUserInRoom(t, conn, user.Id, room.Id)

	err := repo.RegisterByNameInRoom(context.Background(), tx, user.Name, room.Id)
	tx.Close(context.Background())
	assert.True(
		t,
		errors.IsErrorWithCode(err, ErrUserAlreadyRegisteredInRoom),
		"Actual err: %v",
		err,
	)
}

func TestIT_RegistrationRepository_DeleteForRoom(t *testing.T) {
	repo, conn, tx := newTestRegistrationRepositoryAndTransaction(t)
	defer conn.Close(context.Background())
	user := insertTestUser(t, conn)
	room := insertTestRoom(t, conn)
	registerUserInRoom(t, conn, user.Id, room.Id)

	err := repo.DeleteForRoom(context.Background(), tx, room.Id)
	tx.Close(context.Background())
	assert.Nil(t, err, "Actual err: %v", err)

	assertUserExists(t, conn, user.Id)
	assertRoomExists(t, conn, room.Id)
	assertUserNotRegisteredInRoom(t, conn, user.Id, room.Id)
}

func TestIT_RegistrationRepository_DeleteForRoom_OnlyDeletesUserForSpecifiedRoom(t *testing.T) {
	repo, conn, tx := newTestRegistrationRepositoryAndTransaction(t)
	defer conn.Close(context.Background())
	user := insertTestUser(t, conn)
	room1 := insertTestRoom(t, conn)
	room2 := insertTestRoom(t, conn)
	registerUserInRoom(t, conn, user.Id, room1.Id)
	registerUserInRoom(t, conn, user.Id, room2.Id)

	err := repo.DeleteForRoom(context.Background(), tx, room1.Id)
	tx.Close(context.Background())
	assert.Nil(t, err, "Actual err: %v", err)

	assertUserRegisteredInRoom(t, conn, user.Id, room2.Id)
}

func newTestRegistrationRepositoryAndTransaction(t *testing.T) (RegistrationRepository, db.Connection, db.Transaction) {
	conn := newTestConnection(t)
	tx, err := conn.BeginTx(context.Background())
	assert.Nil(t, err, "Actual err: %v", err)
	return NewRegistrationRepository(), conn, tx
}
