package messages

import "github.com/google/uuid"

type Dispatcher interface {
	Broadcast(msg Message)
	BroadcastExcept(id uuid.UUID, msg Message)
	SendTo(id uuid.UUID, msg Message)
}
