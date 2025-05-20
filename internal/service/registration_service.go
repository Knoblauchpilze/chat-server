package service

import (
	"context"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/db"
	"github.com/Knoblauchpilze/chat-server/pkg/repositories"
	"github.com/google/uuid"
)

type RegistrationService interface {
	RegisterUserInRoom(ctx context.Context, user uuid.UUID, room uuid.UUID) error
	UnregisterUserInRoom(ctx context.Context, user uuid.UUID, room uuid.UUID) error
}

type registrationServiceImpl struct {
	conn  db.Connection
	repos repositories.Repositories
}

func NewRegistrationService(conn db.Connection, repos repositories.Repositories) RegistrationService {
	return &registrationServiceImpl{
		conn:  conn,
		repos: repos,
	}
}

func (s *registrationServiceImpl) RegisterUserInRoom(
	ctx context.Context, user uuid.UUID, room uuid.UUID,
) error {
	tx, err := s.conn.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Close(ctx)

	return s.repos.Registration.RegisterInRoom(ctx, tx, user, room)
}

func (s *registrationServiceImpl) UnregisterUserInRoom(
	ctx context.Context, user uuid.UUID, room uuid.UUID,
) error {
	tx, err := s.conn.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Close(ctx)

	err = s.repos.Message.UpdateMessagesOwner(ctx, tx, user, ghostUserName)
	if err != nil {
		return err
	}

	return s.repos.Registration.DeleteFromRoom(ctx, tx, room, user)
}
