package domains

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type (
	RuntimeDomain struct {
		tours   TourRepositoryInterface
		hints   HintRepositoryInterface
		runtime RuntimeRepositoryInterface
	}

	RuntimeDomainInterface interface {
		AppByKey(ctx context.Context, key string) (*App, error)
		CreateApp(ctx context.Context, app *App) (*App, error)
		ListApps(ctx context.Context) ([]*App, error)
		UpdateApp(
			ctx context.Context,
			id uuid.UUID,
			name *string,
			allowedOrigins []string,
		) (*App, error)
		Resolve(
			ctx context.Context,
			req ResolveRequest,
		) (*ResolveResult, error)
		Ping(ctx context.Context) error
	}

	RuntimeRepositoryInterface interface {
		CreateApp(ctx context.Context, app *App) error
		AppByKey(ctx context.Context, key string) (*App, error)
		ListApps(ctx context.Context) ([]*App, error)
		SubjectProgress(
			ctx context.Context,
			appId uuid.UUID,
			subjectId string,
		) (map[uuid.UUID]*Progress, error)
		UpsertProgress(
			ctx context.Context,
			p *Progress,
		) error
		UpdateApp(
			ctx context.Context,
			id uuid.UUID,
			name *string,
			allowedOrigins []string,
		) (*App, error)
		Ping(ctx context.Context) error
	}
)

func NewRuntimeDomain(
	tours TourRepositoryInterface,
	hints HintRepositoryInterface,
	runtime RuntimeRepositoryInterface,
) *RuntimeDomain {
	return &RuntimeDomain{
		tours:   tours,
		hints:   hints,
		runtime: runtime,
	}
}

func (rd *RuntimeDomain) AppByKey(ctx context.Context, key string) (*App, error) {
	return rd.runtime.AppByKey(ctx, key)
}

func (rd *RuntimeDomain) CreateApp(ctx context.Context, app *App) (*App, error) {
	if app.Name == "" {
		return nil, logicErr(ErrTitleRequired, "app validation", http.StatusBadRequest)
	}

	origins, err := ValidateOrigins(app.AllowedOrigins)
	if err != nil {
		return nil, logicErr(err, "app validation", http.StatusBadRequest)
	}

	app.AllowedOrigins = origins

	if app.PublicKey == "" {
		app.PublicKey = "pk_" + uuid.NewString()
	}

	if err := rd.runtime.CreateApp(ctx, app); err != nil {
		return nil, err
	}

	return app, nil
}

func (rd *RuntimeDomain) UpdateApp(
	ctx context.Context,
	id uuid.UUID,
	name *string,
	allowedOrigins []string,
) (*App, error) {
	var normalized []string

	if allowedOrigins != nil {
		var err error
		normalized, err = ValidateOrigins(allowedOrigins)
		if err != nil {
			return nil, logicErr(err, "app update", http.StatusBadRequest)
		}
	}

	return rd.runtime.UpdateApp(ctx, id, name, normalized)
}

func (rd *RuntimeDomain) ListApps(ctx context.Context) ([]*App, error) {
	return rd.runtime.ListApps(ctx)
}

func (rd *RuntimeDomain) Resolve(
	ctx context.Context,
	req ResolveRequest,
) (*ResolveResult, error) {
	if req.SubjectId == "" {
		return nil, logicErr(
			ErrSubjectRequired,
			"resolve",
			http.StatusBadRequest,
		)
	}

	candidates, err := rd.tours.ResolveCandidates(
		ctx,
		req.AppId,
	)
	if err != nil {
		return nil, err
	}

	if len(candidates) == 0 {
		return nil, nil
	}

	progress, err := rd.runtime.SubjectProgress(
		ctx,
		req.AppId,
		req.SubjectId,
	)
	if err != nil {
		return nil, err
	}

	path := PathFromURL(req.Url)

	for _, tour := range candidates {
		if !tour.Enabled || tour.Status != TourPublished {
			continue
		}

		if !MatchPath(
			tour.TargetPath,
			path,
		) {
			continue
		}

		return rd.startScenario(
			ctx,
			req,
			tour,
			progress[tour.Id],
		)
	}

	return nil, nil
}

func (rd *RuntimeDomain) startScenario(
	ctx context.Context,
	req ResolveRequest,
	tour *Tour,
	current *Progress,
) (*ResolveResult, error) {
	hints, err := rd.hints.ListByVersion(
		ctx,
		tour.VersionId,
	)
	if err != nil {
		return nil, err
	}

	tour.Hints = hints

	now := time.Now().UTC()

	switch {
	case current == nil:
		current = &Progress{
			AppId:         req.AppId,
			SubjectId:     req.SubjectId,
			TourId:        tour.Id,
			TourVersionId: tour.VersionId,
			Status:        ProgressInProgress,
			ShowsCount:    1,
			StartedAt:     now,
		}

	case current.TourVersionId != tour.VersionId:
		current = &Progress{
			AppId:         req.AppId,
			SubjectId:     req.SubjectId,
			TourId:        tour.Id,
			TourVersionId: tour.VersionId,
			Status:        ProgressInProgress,
			ShowsCount:    current.ShowsCount + 1,
			StartedAt:     now,
		}

	default:
		current.ShowsCount++
	}

	if current.Status == ProgressInProgress &&
		current.CurrentHintId == nil &&
		len(hints) > 0 {
		current.CurrentHintId = &hints[0].Id
	}

	if err := rd.runtime.UpsertProgress(
		ctx,
		current,
	); err != nil {
		return nil, err
	}

	return &ResolveResult{
		Tour:          tour,
		TourVersionId: tour.VersionId,
		Version:       tour.Version,
		CurrentHintId: current.CurrentHintId,
	}, nil
}

func (rd *RuntimeDomain) Ping(ctx context.Context) error {
	return rd.runtime.Ping(ctx)
}
