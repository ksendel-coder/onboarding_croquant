package response

import (
	"time"

	"github.com/google/uuid"
	"github.com/ksendel-coder/onboarding_croquant/onboarding/internal/domains"
)

type App struct {
	Id             *uuid.UUID `json:"id" format:"uuid"`
	Name           *string    `json:"name" example:"Demo classifieds"`
	PublicKey      *string    `json:"public_key" example:"pk_demo_7f3a91c4b2e85d60"`
	AllowedOrigins []*string  `json:"allowed_origins" example:"http://localhost:3000"`
	CreatedAt      *time.Time `json:"created_at" format:"date-time"`
	ArchivedAt     *time.Time `json:"archived_at,omitempty" format:"date-time"`
}

// Health is the only response here with no domain counterpart: liveness is a
// property of the process, not of the business.
type Health struct {
	Status *string `json:"status" example:"ok"`
	DB     *string `json:"db" example:"ok"`
}

func NewApp(a *domains.App) *App {
	if a == nil {
		return nil
	}

	return &App{
		Id:             a.Id,
		Name:           a.Name,
		PublicKey:      a.PublicKey,
		AllowedOrigins: a.AllowedOrigins,
		CreatedAt:      a.CreatedAt,
		ArchivedAt:     a.ArchivedAt,
	}
}

func NewApps(apps []*domains.App) []*App {
	if apps == nil {
		return nil
	}

	out := make([]*App, 0, len(apps))
	for _, a := range apps {
		out = append(out, NewApp(a))
	}

	return out
}
