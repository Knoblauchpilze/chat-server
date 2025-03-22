package service

import (
	"github.com/Knoblauchpilze/backend-toolkit/pkg/db"
	"github.com/Knoblauchpilze/backend-toolkit/pkg/logger"
	"github.com/Knoblauchpilze/chat-server/pkg/repositories"
)

type Services struct {
	Room RoomService
	Chat ChatService
}

func New(
	conn db.Connection, repos repositories.Repositories, log logger.Logger,
) Services {
	return Services{
		Room: NewRoomService(conn, repos),
		Chat: NewChatService(log),
	}
}
