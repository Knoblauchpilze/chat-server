package communication

import (
	"time"

	"github.com/Knoblauchpilze/chat-server/pkg/persistence"
	"github.com/google/uuid"
)

type RoomDtoRequest struct {
	Name string `json:"name" form:"name"`
}

type RoomDtoResponse struct {
	Id   uuid.UUID `json:"id"`
	Name string    `json:"name"`

	CreatedAt time.Time `json:"created_at"`
}

func FromRoomDtoRequest(room RoomDtoRequest) persistence.Room {
	t := time.Now()
	return persistence.Room{
		Id:   uuid.New(),
		Name: room.Name,

		CreatedAt: t,
		UpdatedAt: t,
	}
}

func ToRoomDtoResponse(room persistence.Room) RoomDtoResponse {
	return RoomDtoResponse{
		Id:   room.Id,
		Name: room.Name,

		CreatedAt: room.CreatedAt,
	}
}
