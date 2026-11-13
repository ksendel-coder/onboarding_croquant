package domains

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	Id           uuid.UUID
	FirstName    string
	LastName     string
	Email        string
	PasswordHash string
	CreatedAt    time.Time
}

type AdminSession struct {
	Id        uuid.UUID
	UserId    uuid.UUID
	TokenHash string
	ExpiresAt time.Time
	CreatedAt time.Time
}
