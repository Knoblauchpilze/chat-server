package service

import (
	"time"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/db"
	"github.com/Knoblauchpilze/backend-toolkit/pkg/logger"
	"github.com/Knoblauchpilze/chat-server/pkg/clients"
	"github.com/Knoblauchpilze/chat-server/pkg/repositories"
)

type Services struct {
	Room RoomService
	User UserService
	Chat ChatService
}

func New(
	connectTimeout time.Duration,
	conn db.Connection,
	repos repositories.Repositories,
	log logger.Logger,
) Services {
	return Services{
		Room: NewRoomService(conn, repos),
		User: NewUserService(conn, repos),
		Chat: NewChatService(
			clients.Handshake,
			connectTimeout,
			log,
		),
	}
}
