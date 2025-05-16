package service

import (
	"context"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/db"
	"github.com/Knoblauchpilze/backend-toolkit/pkg/errors"
	"github.com/Knoblauchpilze/chat-server/pkg/communication"
	"github.com/Knoblauchpilze/chat-server/pkg/repositories"
	"github.com/google/uuid"
)

type RoomService interface {
	Create(ctx context.Context, roomDto communication.RoomDtoRequest) (communication.RoomDtoResponse, error)
	Get(ctx context.Context, id uuid.UUID) (communication.RoomDtoResponse, error)
	ListUserForRoom(ctx context.Context, room uuid.UUID) ([]communication.UserDtoResponse, error)
	ListMessageForRoom(ctx context.Context, room uuid.UUID) ([]communication.MessageDtoResponse, error)
	RegisterUserInRoom(ctx context.Context, user uuid.UUID, room uuid.UUID) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type roomServiceImpl struct {
	conn db.Connection

	roomRepo    repositories.RoomRepository
	userRepo    repositories.UserRepository
	messageRepo repositories.MessageRepository
}

func NewRoomService(conn db.Connection, repos repositories.Repositories) RoomService {
	return &roomServiceImpl{
		conn:        conn,
		roomRepo:    repos.Room,
		userRepo:    repos.User,
		messageRepo: repos.Message,
	}
}

func (s *roomServiceImpl) Create(
	ctx context.Context, roomDto communication.RoomDtoRequest,
) (communication.RoomDtoResponse, error) {
	room := communication.FromRoomDtoRequest(roomDto)

	if room.Name == "" {
		return communication.RoomDtoResponse{}, errors.NewCode(ErrInvalidName)
	}

	tx, err := s.conn.BeginTx(ctx)
	if err != nil {
		return communication.RoomDtoResponse{}, err
	}
	defer tx.Close(ctx)

	createdRoom, err := s.roomRepo.Create(ctx, tx, room)
	if err != nil {
		return communication.RoomDtoResponse{}, err
	}

	err = s.roomRepo.RegisterUserByNameInRoom(
		ctx, tx, ghostUserName, room.Id,
	)
	if err != nil {
		return communication.RoomDtoResponse{}, err
	}

	out := communication.ToRoomDtoResponse(createdRoom)
	return out, nil
}

func (s *roomServiceImpl) Get(
	ctx context.Context, id uuid.UUID,
) (communication.RoomDtoResponse, error) {
	room, err := s.roomRepo.Get(ctx, id)
	if err != nil {
		return communication.RoomDtoResponse{}, err
	}

	out := communication.ToRoomDtoResponse(room)
	return out, nil
}

func (s *roomServiceImpl) ListUserForRoom(
	ctx context.Context, room uuid.UUID,
) ([]communication.UserDtoResponse, error) {
	users, err := s.userRepo.ListForRoom(ctx, room)
	if err != nil {
		return []communication.UserDtoResponse{}, err
	}

	out := make([]communication.UserDtoResponse, 0)
	for _, user := range users {
		dto := communication.ToUserDtoResponse(user)
		out = append(out, dto)
	}

	return out, nil
}

func (s *roomServiceImpl) ListMessageForRoom(
	ctx context.Context, room uuid.UUID,
) ([]communication.MessageDtoResponse, error) {
	messages, err := s.messageRepo.ListForRoom(ctx, room)
	if err != nil {
		return []communication.MessageDtoResponse{}, err
	}

	out := make([]communication.MessageDtoResponse, 0)
	for _, message := range messages {
		dto := communication.ToMessageDtoResponse(message)
		out = append(out, dto)
	}

	return out, nil
}

func (s *roomServiceImpl) RegisterUserInRoom(
	ctx context.Context, user uuid.UUID, room uuid.UUID,
) error {
	tx, err := s.conn.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Close(ctx)

	return s.roomRepo.RegisterUserInRoom(ctx, tx, user, room)
}

func (s *roomServiceImpl) Delete(
	ctx context.Context, id uuid.UUID,
) error {
	tx, err := s.conn.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Close(ctx)

	err = s.roomRepo.Delete(ctx, tx, id)
	if err != nil {
		return err
	}

	return nil
}
