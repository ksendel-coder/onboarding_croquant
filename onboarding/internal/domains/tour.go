package domains

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

type TourDomain struct {
	tours TourRepositoryInterface
}

type TourDomainInterface interface {
	Create(
		ctx context.Context,
		appId uuid.UUID,
		t *Tour,
	) (*Tour, error)

	List(
		ctx context.Context,
		appId uuid.UUID,
	) ([]*Tour, error)

	UpdateMeta(
		ctx context.Context,
		tourId uuid.UUID,
		p TourMetaPatch,
	) (*Tour, error)

	Archive(
		ctx context.Context,
		tourId uuid.UUID,
	) error
}

type TourRepositoryInterface interface {
	CreateWithDraft(
		ctx context.Context,
		t *Tour,
	) (*Tour, error)

	GetById(
		ctx context.Context,
		id uuid.UUID,
	) (*Tour, error)

	List(
		ctx context.Context,
		appId uuid.UUID,
	) ([]*Tour, error)

	UpdateMeta(
		ctx context.Context,
		id uuid.UUID,
		p TourMetaPatch,
	) (*Tour, error)

	Archive(
		ctx context.Context,
		id uuid.UUID,
	) error
}

type TourMetaPatch struct {
	Title       *string
	Description *string
	Enabled     *bool
	Priority    *int
}

func NewTourDomain(
	tours TourRepositoryInterface,
) *TourDomain {
	return &TourDomain{
		tours: tours,
	}
}

func (td *TourDomain) Create(
	ctx context.Context,
	appId uuid.UUID,
	t *Tour,
) (*Tour, error) {
	if err := validateTour(t); err != nil {
		return nil, err
	}

	t.AppId = appId
	t.Status = TourDraft

	if t.TriggerType == "" {
		t.TriggerType = TriggerOnLoad
	}

	if t.Audience == nil {
		t.Audience = &Audience{}
	}

	return td.tours.CreateWithDraft(ctx, t)
}

func (td *TourDomain) List(
	ctx context.Context,
	appId uuid.UUID,
) ([]*Tour, error) {
	return td.tours.List(ctx, appId)
}

func (td *TourDomain) UpdateMeta(
	ctx context.Context,
	tourId uuid.UUID,
	p TourMetaPatch,
) (*Tour, error) {
	if p.Title != nil && *p.Title == "" {
		return nil, logicErr(
			ErrTitleRequired,
			"tour update",
			http.StatusBadRequest,
		)
	}

	if p.Priority != nil && *p.Priority < 0 {
		return nil, logicErr(
			ErrPriorityNegative,
			"tour update",
			http.StatusBadRequest,
		)
	}

	return td.tours.UpdateMeta(
		ctx,
		tourId,
		p,
	)
}

func (td *TourDomain) Archive(
	ctx context.Context,
	tourId uuid.UUID,
) error {
	return td.tours.Archive(
		ctx,
		tourId,
	)
}

func validateTour(t *Tour) error {
	var err error

	switch {
	case t.Title == "":
		err = ErrTitleRequired

	case t.TargetPath == "":
		err = ErrTargetPathRequired

	case t.Priority < 0:
		err = ErrPriorityNegative

	default:
		return nil
	}

	return logicErr(
		err,
		"tour validation",
		http.StatusBadRequest,
	)
}
