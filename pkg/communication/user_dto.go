package communication

import (
	"time"

	"github.com/Knoblauchpilze/chat-server/pkg/persistence"
	"github.com/google/uuid"
)

type UserDtoRequest struct {
	Name    string    `json:"name" form:"name"`
	ApiUser uuid.UUID `json:"api_user"`
}

type UserDtoResponse struct {
	Id      uuid.UUID `json:"id"`
	Name    string    `json:"name"`
	ApiUser uuid.UUID `json:"api_user"`

	CreatedAt time.Time `json:"created_at"`
}

func FromUserDtoRequest(user UserDtoRequest) persistence.User {
	t := time.Now()
	return persistence.User{
		Id:      uuid.New(),
		Name:    user.Name,
		ApiUser: user.ApiUser,

		CreatedAt: t,
		UpdatedAt: t,
	}
}

func ToUserDtoResponse(user persistence.User) UserDtoResponse {
	return UserDtoResponse{
		Id:      user.Id,
		Name:    user.Name,
		ApiUser: user.ApiUser,

		CreatedAt: user.CreatedAt,
	}
}
