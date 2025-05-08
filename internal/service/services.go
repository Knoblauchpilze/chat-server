package service

import (
	"github.com/Knoblauchpilze/backend-toolkit/pkg/db"
	"github.com/Knoblauchpilze/backend-toolkit/pkg/logger"
	"github.com/Knoblauchpilze/chat-server/pkg/clients"
	"github.com/Knoblauchpilze/chat-server/pkg/messages"
	"github.com/Knoblauchpilze/chat-server/pkg/repositories"
)

type Services struct {
	Room    RoomService
	User    UserService
	Message MessageService
}

func New(
	conn db.Connection,
	repos repositories.Repositories,
	processor messages.Processor,
	manager clients.Manager,
	log logger.Logger,
) Services {
	return Services{
		Room:    NewRoomService(conn, repos),
		User:    NewUserService(conn, repos),
		Message: NewMessageService(conn, processor, manager),
	}
}
