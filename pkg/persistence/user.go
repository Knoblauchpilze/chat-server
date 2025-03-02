package persistence

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	Id      uuid.UUID
	Name    string
	ApiUser uuid.UUID

	CreatedAt time.Time
	UpdatedAt time.Time

	Version int
}
