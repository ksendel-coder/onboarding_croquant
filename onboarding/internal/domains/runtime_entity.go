package domains

import (
	"time"

	"github.com/google/uuid"
)

type App struct {
	Id             uuid.UUID
	Name           string
	PublicKey      string
	AllowedOrigins []string
	CreatedAt      time.Time
	ArchivedAt     *time.Time
}

type ResolveRequest struct {
	AppId     uuid.UUID
	Url       string
	SubjectId string
	SessionId string
	Props     map[string]any
}
