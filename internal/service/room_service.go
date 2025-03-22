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
	Delete(ctx context.Context, id uuid.UUID) error
}

type roomServiceImpl struct {
	conn db.Connection

	roomRepo repositories.RoomRepository
}

func NewRoomService(conn db.Connection, repos repositories.Repositories) RoomService {
	return &roomServiceImpl{
		conn:     conn,
		roomRepo: repos.Room,
	}
}

func (s *roomServiceImpl) Create(
	ctx context.Context, roomDto communication.RoomDtoRequest,
) (communication.RoomDtoResponse, error) {
	room := communication.FromRoomDtoRequest(roomDto)

	if room.Name == "" {
		return communication.RoomDtoResponse{}, errors.NewCode(InvalidName)
	}

	createdRoom, err := s.roomRepo.Create(ctx, room)
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
