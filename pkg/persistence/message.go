package persistence

import (
	"time"

	"github.com/google/uuid"
)

type Message struct {
	Id        uuid.UUID
	User      uuid.UUID
	Room      uuid.UUID
	Message   string
	CreatedAt time.Time
}
