package response

import (
	"time"

	"github.com/google/uuid"
	"github.com/ksendel-coder/onboarding_croquant/onboarding/internal/domains"
)

type Hint struct {
	Id            uuid.UUID `json:"id"`
	TourVersionId uuid.UUID `json:"tour_version_id"`
	Step          int       `json:"step"`
	Title         string    `json:"title"`
	Content       string    `json:"content"`
	Selector      string    `json:"selector"`
	Placement     string    `json:"placement"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func NewHint(
	h *domains.Hint,
) *Hint {
	if h == nil {
		return nil
	}

	return &Hint{
		Id:            h.Id,
		TourVersionId: h.TourVersionId,
		Step:          h.Step,
		Title:         h.Title,
		Content:       h.Content,
		Selector:      h.Selector,
		Placement:     string(h.Placement),
		CreatedAt:     h.CreatedAt,
		UpdatedAt:     h.UpdatedAt,
	}
}

func NewHints(
	hints []domains.Hint,
) []Hint {
	if hints == nil {
		return nil
	}

	out := make(
		[]Hint,
		0,
		len(hints),
	)

	for _, h := range hints {
		out = append(
			out,
			*NewHint(&h),
		)
	}

	return out
}
