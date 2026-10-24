package domains

import (
	"time"

	"github.com/google/uuid"
)

type Placement string

const (
	PlacementTop      Placement = "top"
	PlacementBottom   Placement = "bottom"
	PlacementCenter   Placement = "center"
	PlacementLeftTop  Placement = "left-top"
	PlacementRightTop Placement = "right-top"
	PlacementLeft     Placement = "left"
	PlacementRight    Placement = "right"
)

func (p Placement) Valid() bool {
	switch p {
	case PlacementTop,
		PlacementBottom,
		PlacementCenter,
		PlacementLeftTop,
		PlacementRightTop,
		PlacementLeft,
		PlacementRight:
		return true
	}

	return false
}

type Hint struct {
	Id            uuid.UUID
	TourVersionId uuid.UUID
	Step          int
	Title         string
	Content       string
	Selector      string
	Placement     Placement
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type HintPatch struct {
	Selector  *string
	Title     *string
	Content   *string
	Placement *Placement
}
