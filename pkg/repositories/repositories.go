package repositories

import "github.com/Knoblauchpilze/backend-toolkit/pkg/db"

type Repositories struct {
	User    UserRepository
	Room    RoomRepository
	Message MessageRepository
}

func New(conn db.Connection) Repositories {
	return Repositories{
		User:    NewUserRepository(conn),
		Room:    NewRoomRepository(conn),
		Message: NewMessageRepository(conn),
	}
}
