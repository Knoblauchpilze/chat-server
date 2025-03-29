package persistence

import (
	"time"

	"github.com/google/uuid"
)

type Message struct {
	Id        uuid.UUID
	ChatUser  uuid.UUID
	Room      uuid.UUID
	Message   string
	CreatedAt time.Time
}
