package domains

import (
	"context"
	"net/http"
	"net/url"
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
		CreateApp(
			ctx context.Context,
			userId uuid.UUID,
			app *App,
		) (*App, error)
		ListApps(
			ctx context.Context,
			userId uuid.UUID,
		) ([]*App, error)
		UpdateApp(
			ctx context.Context,
			userId uuid.UUID,
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
		AppByKey(ctx context.Context, key string) (*App, error)
		CreateApp(
			ctx context.Context,
			userId uuid.UUID,
			app *App,
		) error
		ListApps(
			ctx context.Context,
			userId uuid.UUID,
		) ([]*App, error)
		SubjectProgress(
			ctx context.Context,
			appId uuid.UUID,
			subjectId string,
		) (map[uuid.UUID]*Progress, error)
		UpsertProgress(
			ctx context.Context,
			p *Progress,
		) (bool, error)
		UpdateApp(
			ctx context.Context,
			userId uuid.UUID,
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

func (rd *RuntimeDomain) CreateApp(
	ctx context.Context,
	userId uuid.UUID,
	app *App,
) (*App, error) {
	if app.Name == "" {
		return nil, logicErr(
			ErrTitleRequired,
			"app validation",
			http.StatusBadRequest,
		)
	}

	origins, err := ValidateOrigins(
		app.AllowedOrigins,
	)
	if err != nil {
		return nil, logicErr(
			err,
			"app validation",
			http.StatusBadRequest,
		)
	}

	app.AllowedOrigins = origins

	if app.PublicKey == "" {
		app.PublicKey = "pk_" + uuid.NewString()
	}

	if err := rd.runtime.CreateApp(
		ctx,
		userId,
		app,
	); err != nil {
		return nil, err
	}

	return app, nil
}

func (rd *RuntimeDomain) UpdateApp(
	ctx context.Context,
	userId uuid.UUID,
	id uuid.UUID,
	name *string,
	allowedOrigins []string,
) (*App, error) {
	var normalized []string

	if allowedOrigins != nil {
		var err error

		normalized, err = ValidateOrigins(
			allowedOrigins,
		)
		if err != nil {
			return nil, logicErr(
				err,
				"app update",
				http.StatusBadRequest,
			)
		}
	}

	return rd.runtime.UpdateApp(
		ctx,
		userId,
		id,
		name,
		normalized,
	)
}

func (rd *RuntimeDomain) ListApps(
	ctx context.Context,
	userId uuid.UUID,
) ([]*App, error) {
	return rd.runtime.ListApps(
		ctx,
		userId,
	)
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
	// Resume before starting another tour, including when a step changes pages.
	for _, tour := range candidates {
		current := progress[tour.Id]
		if !tour.Enabled || tour.Status != TourPublished || current == nil ||
			current.Status != ProgressInProgress || current.TourVersionId != tour.VersionId {
			continue
		}
		resumed, err := rd.startScenario(ctx, req, tour, current, path)
		if err != nil || resumed != nil {
			return resumed, err
		}
	}

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

		current := progress[tour.Id]
		if current != nil {
			// Если пользователь проходит старую опубликованную версию,
			// не перезаписываем её новой версией из /resolve.
			// События MVP-06 сверяются с tour_progress.tour_version_id.
			if current.Status == ProgressInProgress {
				continue
			}
		}

		if tour.Audience != nil {
			if tour.Audience.OnlyNew && current != nil {
				continue
			}
			if tour.Audience.ShowOnce && current != nil && current.ShowsCount > 0 {
				continue
			}
			if tour.Audience.MaxShows > 0 && current != nil && current.ShowsCount >= tour.Audience.MaxShows {
				continue
			}
		}

		return rd.startScenario(
			ctx,
			req,
			tour,
			current,
			path,
		)
	}

	return nil, nil
}

func (rd *RuntimeDomain) startScenario(
	ctx context.Context,
	req ResolveRequest,
	tour *Tour,
	current *Progress,
	path string,
) (*ResolveResult, error) {
	hints, err := rd.hints.ListByVersion(
		ctx,
		tour.VersionId,
	)
	if err != nil {
		return nil, err
	}

	newProgress := current == nil || current.Status != ProgressInProgress
	if !newProgress && !MatchPath(tour.TargetPath, path) {
		var pagePath string
		for _, hint := range hints {
			if hint.PagePath != "" {
				pagePath, _ = url.PathUnescape(hint.PagePath)
			}
			if current.CurrentHintId != nil && hint.Id == *current.CurrentHintId {
				break
			}
		}
		if !MatchPath(pagePath, path) {
			return nil, nil
		}
	}

	tour.Hints = hints
	tour.StepsCount = len(hints)
	if newProgress {
		shows := 1
		if current != nil {
			shows += current.ShowsCount
		}
		attemptId := uuid.New()
		next := &Progress{
			AppId:         req.AppId,
			SubjectId:     req.SubjectId,
			TourId:        tour.Id,
			TourVersionId: tour.VersionId,
			Status:        ProgressInProgress,
			ShowsCount:    shows,
			StartedAt:     time.Now().UTC(),
			AttemptId:     &attemptId,
		}
		if len(hints) > 0 {
			next.CurrentHintId = &hints[0].Id
		}
		written, err := rd.runtime.UpsertProgress(ctx, next)
		if err != nil {
			return nil, err
		}
		current = next
		if !written {
			progress, err := rd.runtime.SubjectProgress(ctx, req.AppId, req.SubjectId)
			if err != nil {
				return nil, err
			}
			current = progress[tour.Id]
			if current == nil || current.Status != ProgressInProgress || current.TourVersionId != tour.VersionId {
				return nil, nil
			}
		}
	}

	attemptId := current.AttemptId
	if attemptId == nil {
		id := uuid.NewSHA1(tour.VersionId, []byte(req.SubjectId))
		attemptId = &id
	}
	return &ResolveResult{
		Tour:          tour,
		TourVersionId: tour.VersionId,
		Version:       tour.Version,
		CurrentHintId: current.CurrentHintId,
		AttemptId:     attemptId,
	}, nil
}

func (rd *RuntimeDomain) Ping(ctx context.Context) error {
	return rd.runtime.Ping(ctx)
}
