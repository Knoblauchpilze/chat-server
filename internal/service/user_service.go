package service

import (
	"context"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/db"
	"github.com/Knoblauchpilze/backend-toolkit/pkg/errors"
	"github.com/Knoblauchpilze/chat-server/pkg/communication"
	"github.com/Knoblauchpilze/chat-server/pkg/repositories"
	"github.com/google/uuid"
)

const generalRoomName = "general"

type UserService interface {
	Create(ctx context.Context, userDto communication.UserDtoRequest) (communication.UserDtoResponse, error)
	Get(ctx context.Context, id uuid.UUID) (communication.UserDtoResponse, error)
	GetByName(ctx context.Context, name string) (communication.UserDtoResponse, error)
	ListForUser(ctx context.Context, user uuid.UUID) ([]communication.RoomDtoResponse, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type userServiceImpl struct {
	conn db.Connection

	roomRepo repositories.RoomRepository
	userRepo repositories.UserRepository
}

func NewUserService(conn db.Connection, repos repositories.Repositories) UserService {
	return &userServiceImpl{
		conn:     conn,
		roomRepo: repos.Room,
		userRepo: repos.User,
	}
}

func (s *userServiceImpl) Create(
	ctx context.Context, userDto communication.UserDtoRequest,
) (communication.UserDtoResponse, error) {
	user := communication.FromUserDtoRequest(userDto)

	if user.Name == "" {
		return communication.UserDtoResponse{}, errors.NewCode(ErrInvalidName)
	}

	tx, err := s.conn.BeginTx(ctx)
	if err != nil {
		return communication.UserDtoResponse{}, err
	}
	defer tx.Close(ctx)

	createdUser, err := s.userRepo.Create(ctx, tx, user)
	if err != nil {
		return communication.UserDtoResponse{}, err
	}

	err = s.roomRepo.RegisterUserInRoomByName(
		ctx, tx, createdUser.Id, generalRoomName,
	)
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

func (s *userServiceImpl) GetByName(
	ctx context.Context, name string,
) (communication.UserDtoResponse, error) {
	user, err := s.userRepo.GetByName(ctx, name)
	if err != nil {
		return communication.UserDtoResponse{}, err
	}

	out := communication.ToUserDtoResponse(user)
	return out, nil
}

func (s *userServiceImpl) ListForUser(ctx context.Context, user uuid.UUID) ([]communication.RoomDtoResponse, error) {
	rooms, err := s.roomRepo.ListForUser(ctx, user)
	if err != nil {
		return []communication.RoomDtoResponse{}, err
	}

	out := make([]communication.RoomDtoResponse, 0)
	for _, room := range rooms {
		dto := communication.ToRoomDtoResponse(room)
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

	err = s.roomRepo.DeleteUserFromRoomByName(
		ctx, tx, id, generalRoomName,
	)
	if err != nil {
		return err
	}

	err = s.userRepo.Delete(ctx, tx, id)
	if err != nil {
		return err
	}

	return nil
}
