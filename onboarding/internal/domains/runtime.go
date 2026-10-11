package domains

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

type (
	RuntimeDomain struct {
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
		Ping(ctx context.Context) error
	}

	RuntimeRepositoryInterface interface {
		CreateApp(ctx context.Context, app *App) error
		AppByKey(ctx context.Context, key string) (*App, error)
		ListApps(ctx context.Context) ([]*App, error)
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
	runtime RuntimeRepositoryInterface,
) *RuntimeDomain {
	return &RuntimeDomain{runtime: runtime}
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

func (rd *RuntimeDomain) Ping(ctx context.Context) error {
	return rd.runtime.Ping(ctx)
}
