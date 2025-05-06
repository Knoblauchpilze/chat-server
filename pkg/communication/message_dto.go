package communication

import (
	"time"

	"github.com/Knoblauchpilze/chat-server/pkg/persistence"
	"github.com/google/uuid"
)

type MessageDtoRequest struct {
	User    uuid.UUID `json:"user"`
	Room    uuid.UUID `json:"room"`
	Message string    `json:"message"`
}

type MessageDtoResponse struct {
	Id      uuid.UUID `json:"id"`
	User    uuid.UUID `json:"user"`
	Room    uuid.UUID `json:"room"`
	Message string    `json:"message"`

	CreatedAt time.Time `json:"created_at"`
}

func FromMessageDtoRequest(message MessageDtoRequest) persistence.Message {
	t := time.Now().UTC()
	return persistence.Message{
		Id:       uuid.New(),
		ChatUser: message.User,
		Room:     message.Room,
		Message:  message.Message,

		CreatedAt: t,
	}
}

func ToMessageDtoResponse(message persistence.Message) MessageDtoResponse {
	return MessageDtoResponse{
		Id:      message.Id,
		User:    message.ChatUser,
		Room:    message.Room,
		Message: message.Message,

		CreatedAt: message.CreatedAt,
	}
}
