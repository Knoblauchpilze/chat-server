package messages

import (
	"github.com/Knoblauchpilze/chat-server/pkg/persistence"
	"github.com/google/uuid"
)

type Dispatcher interface {
	Broadcast(msg persistence.Message)
	BroadcastExcept(id uuid.UUID, msg persistence.Message)
	SendTo(id uuid.UUID, msg persistence.Message)
}
