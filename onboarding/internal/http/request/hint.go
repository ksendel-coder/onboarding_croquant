package request

import (
	"github.com/google/uuid"
	"github.com/ksendel-coder/onboarding_croquant/onboarding/internal/domains"
)

type Hint struct {
	Title     string `json:"title"`
	Content   string `json:"content"`
	Selector  string `json:"selector"`
	Placement string `json:"placement"`
}

type HintPatch struct {
	Selector  *string `json:"selector"`
	Title     *string `json:"title"`
	Content   *string `json:"content"`
	Placement *string `json:"placement"`
}

type ReorderRequest struct {
	HintIds []uuid.UUID `json:"hint_ids" binding:"required"`
}

func (h *Hint) ToDomain() *domains.Hint {
	if h == nil {
		return nil
	}

	return &domains.Hint{
		Title:     h.Title,
		Content:   h.Content,
		Selector:  h.Selector,
		Placement: domains.Placement(h.Placement),
	}
}

func (p *HintPatch) ToDomain() domains.HintPatch {
	if p == nil {
		return domains.HintPatch{}
	}

	patch := domains.HintPatch{
		Selector: p.Selector,
		Title:    p.Title,
		Content:  p.Content,
	}

	if p.Placement != nil {
		placement := domains.Placement(
			*p.Placement,
		)

		patch.Placement = &placement
	}

	return patch
}
