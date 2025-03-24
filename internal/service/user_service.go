package service

import (
	"context"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/db"
	"github.com/Knoblauchpilze/backend-toolkit/pkg/errors"
	"github.com/Knoblauchpilze/chat-server/pkg/communication"
	"github.com/Knoblauchpilze/chat-server/pkg/repositories"
	"github.com/google/uuid"
)

type UserService interface {
	Create(ctx context.Context, userDto communication.UserDtoRequest) (communication.UserDtoResponse, error)
	Get(ctx context.Context, id uuid.UUID) (communication.UserDtoResponse, error)
	ListForRoom(ctx context.Context, room uuid.UUID) ([]communication.UserDtoResponse, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type userServiceImpl struct {
	conn db.Connection

	userRepo repositories.UserRepository
}

func NewUserService(conn db.Connection, repos repositories.Repositories) UserService {
	return &userServiceImpl{
		conn:     conn,
		userRepo: repos.User,
	}
}

func (s *userServiceImpl) Create(
	ctx context.Context, userDto communication.UserDtoRequest,
) (communication.UserDtoResponse, error) {
	user := communication.FromUserDtoRequest(userDto)

	if user.Name == "" {
		return communication.UserDtoResponse{}, errors.NewCode(InvalidName)
	}

	createdUser, err := s.userRepo.Create(ctx, user)
	if err != nil {
		return communication.UserDtoResponse{}, err
	}

	out := communication.ToUserDtoResponse(createdUser)
	return out, nil
}

func (s *userServiceImpl) Get(
	ctx context.Context, id uuid.UUID,
) (communication.UserDtoResponse, error) {
	user, err := s.userRepo.Get(ctx, id)
	if err != nil {
		return communication.UserDtoResponse{}, err
	}

	out := communication.ToUserDtoResponse(user)
	return out, nil
}

// TODO: Add tests for this function
func (s *userServiceImpl) ListForRoom(
	ctx context.Context, room uuid.UUID,
) ([]communication.UserDtoResponse, error) {
	users, err := s.userRepo.ListForRoom(ctx, room)
	if err != nil {
		return []communication.UserDtoResponse{}, err
	}

	var out []communication.UserDtoResponse
	for _, user := range users {
		dto := communication.ToUserDtoResponse(user)
		out = append(out, dto)
	}

	return out, nil
}

func (s *userServiceImpl) Delete(
	ctx context.Context, id uuid.UUID,
) error {
	tx, err := s.conn.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Close(ctx)

	err = s.userRepo.Delete(ctx, tx, id)
	if err != nil {
		return err
	}

	return nil
}
