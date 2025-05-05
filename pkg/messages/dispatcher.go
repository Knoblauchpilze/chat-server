package messages

import "github.com/google/uuid"

type Dispatcher interface {
	// TODO: This should be using communication.MessageResponseDto
	Broadcast(msg Message)
	BroadcastExcept(id uuid.UUID, msg Message)
	SendTo(id uuid.UUID, msg Message)
}
