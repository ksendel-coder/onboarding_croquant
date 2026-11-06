package domains

import (
	"context"
	"errors"
	"net/http"
	"strings"

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

	Publish(
		ctx context.Context,
		tourId uuid.UUID,
	) (*PublishResult, error)

	UpdateSettings(
		ctx context.Context,
		tourId uuid.UUID,
		p TourSettingsPatch,
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

	ResolveCandidates(
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

	VersionByStatus(
		ctx context.Context,
		tourId uuid.UUID,
		status TourStatus,
	) (*TourVersion, error)

	Publish(
		ctx context.Context,
		tourId uuid.UUID,
	) (*PublishResult, error)

	UpdateSettings(
		ctx context.Context,
		tourId uuid.UUID,
		p TourSettingsPatch,
	) error
}

type TourMetaPatch struct {
	Title       *string
	Description *string
	Enabled     *bool
	Priority    *int
}

type TourSettingsPatch struct {
	TargetPath       *string
	TriggerType      *TriggerType
	TriggerConfig    *TriggerConfig
	AudienceShowOnce *bool
	AudienceMaxShows *int
	AudienceOnlyNew  *bool
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

func (td *TourDomain) Publish(ctx context.Context, tourId uuid.UUID) (*PublishResult, error) {
	if tourId == uuid.Nil {
		return nil, logicErr(ErrTourNotFound, "tour publish", http.StatusBadRequest)
	}
	return td.tours.Publish(ctx, tourId)
}

func (td *TourDomain) UpdateSettings(
	ctx context.Context,
	tourId uuid.UUID,
	p TourSettingsPatch,
) error {
	if p.TargetPath != nil && strings.TrimSpace(*p.TargetPath) == "" {
		return logicErr(errors.New("target_path is required"), "tour settings", http.StatusBadRequest)
	}
	if p.AudienceMaxShows != nil && *p.AudienceMaxShows < 0 {
		return logicErr(errors.New("audience_max_shows must be >= 0"), "tour settings", http.StatusBadRequest)
	}
	if p.TriggerType != nil {
		switch *p.TriggerType {
		case TriggerOnLoad, TriggerDelay, TriggerExitIntent, TriggerManual,
			TriggerScrollDepth, TriggerInactivity, TriggerElementVisible:
		default:
			return logicErr(errors.New("invalid trigger_type"), "tour settings", http.StatusBadRequest)
		}
	}
	if p.TriggerConfig != nil {
		config := p.TriggerConfig
		if config.DelayMs != nil && (*config.DelayMs < 0 || *config.DelayMs > 600000) {
			return logicErr(errors.New("invalid delay_ms"), "tour settings", http.StatusBadRequest)
		}
		if config.ScrollDepth != nil && (*config.ScrollDepth < 0 || *config.ScrollDepth > 100) {
			return logicErr(errors.New("invalid scroll_depth"), "tour settings", http.StatusBadRequest)
		}
		if config.InactivitySecs != nil && (*config.InactivitySecs < 0 || *config.InactivitySecs > 86400) {
			return logicErr(errors.New("invalid inactivity_secs"), "tour settings", http.StatusBadRequest)
		}
		if config.ElementSelector != nil && strings.TrimSpace(*config.ElementSelector) == "" {
			return logicErr(errors.New("empty element_selector"), "tour settings", http.StatusBadRequest)
		}
	}

	return td.tours.UpdateSettings(ctx, tourId, p)
}
