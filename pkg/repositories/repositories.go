package repositories

import "github.com/Knoblauchpilze/backend-toolkit/pkg/db"

type Repositories struct {
	Message MessageRepository
	Room    RoomRepository
	User    UserRepository
}

func New(conn db.Connection) Repositories {
	return Repositories{
		Message: NewMessageRepository(conn),
		Room:    NewRoomRepository(conn),
		User:    NewUserRepository(conn),
	}
}
