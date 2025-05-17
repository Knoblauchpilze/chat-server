package repositories

import "github.com/Knoblauchpilze/backend-toolkit/pkg/db"

type Repositories struct {
	Message      MessageRepository
	Registration RegistrationRepository
	Room         RoomRepository
	User         UserRepository
}

func New(conn db.Connection) Repositories {
	return Repositories{
		Message:      NewMessageRepository(conn),
		Registration: NewRegistrationRepository(),
		Room:         NewRoomRepository(conn),
		User:         NewUserRepository(conn),
	}
}
