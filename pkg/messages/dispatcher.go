package messages

import (
	"github.com/Knoblauchpilze/chat-server/pkg/persistence"
	"github.com/google/uuid"
)

type MessageDispatcher interface {
	Broadcast(msg Message)
	BroadcastExcept(id uuid.UUID, msg Message)
	SendTo(id uuid.UUID, msg Message)
}

type Dispatcher interface {
	Broadcast(msg persistence.Message)
	BroadcastExcept(id uuid.UUID, msg persistence.Message)
	SendTo(id uuid.UUID, msg persistence.Message)
}
