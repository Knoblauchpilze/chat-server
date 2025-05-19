package service

type Services struct {
	Registration RegistrationService
	Room         RoomService
	User         UserService
	Message      MessageService
}
