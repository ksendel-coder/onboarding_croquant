package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ksendel-coder/onboarding_croquant/onboarding/internal/domains"
	"github.com/ksendel-coder/onboarding_croquant/onboarding/internal/http/request"
	"github.com/ksendel-coder/onboarding_croquant/onboarding/internal/http/response"
	"github.com/ksendel-coder/onboarding_croquant/platform/errs"
)

type (
	RuntimeController struct {
		domain domains.RuntimeDomainInterface
	}

	RuntimeControllerInterface interface {
		Health(ctx *gin.Context)
		CreateApp(ctx *gin.Context)
		ListApps(ctx *gin.Context)
		UpdateApp(ctx *gin.Context)
		Resolve(ctx *gin.Context)
	}
)

func NewRuntimeController(domain domains.RuntimeDomainInterface) *RuntimeController {
	return &RuntimeController{domain: domain}
}

// Health godoc
//
// @Summary        Health check
// @Description    Reports whether the process and its database are both usable
// @Tags           Runtime
// @Produce        json
// @Success        200 {object} response.Health
// @Failure        503 {object} response.Health
// @Router         /health [get]
func (rc *RuntimeController) Health(ctx *gin.Context) {
	if err := rc.domain.Ping(ctx.Request.Context()); err != nil {
		ctx.JSON(http.StatusServiceUnavailable, response.Health{Status: "degraded", DB: "down"})
		return
	}

	ctx.JSON(http.StatusOK, response.Health{Status: "ok", DB: "ok"})
}

// CreateApp godoc
//
// @Summary        Register a consumer application
// @Description    Issues the public key that goes into the SDK script tag
// @Tags           Apps
// @Accept         json
// @Produce        json
// @Param          body body request.CreateAppRequest true "Application"
// @Success        201 {object} response.App
// @Failure        400 {object} errs.HTTPError
// @Router         /apps [post]
func (rc *RuntimeController) CreateApp(ctx *gin.Context) {
	body := new(request.CreateAppRequest)

	if err := ctx.ShouldBindJSON(body); err != nil {
		ctx.Error(&errs.BindingError{Err: err, Zone: errs.Body, Code: http.StatusBadRequest})
		return
	}

	created, err := rc.domain.CreateApp(ctx.Request.Context(), body.ToDomain())
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.Header("Location", "/v1/apps/"+created.Id.String())
	ctx.JSON(http.StatusCreated, response.NewApp(created))
}

// ListApps godoc
//
// @Summary        List consumer applications
// @Tags           Apps
// @Produce        json
// @Success        200 {array} response.App
// @Router         /apps [get]
func (rc *RuntimeController) ListApps(ctx *gin.Context) {
	apps, err := rc.domain.ListApps(ctx.Request.Context())
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, response.NewApps(apps))
}

// Resolve godoc
//
// @Summary        Resolve an onboarding scenario
// @Description    Returns the first matching published scenario and its steps
// @Tags           Runtime
// @Accept         json
// @Produce        json
// @Param          X-App-Key header string true "Application public key"
// @Param          body body request.ResolveRequest true "Resolve context"
// @Success        200 {object} response.ResolveResult
// @Success        204
// @Failure        400 {object} errs.HTTPError
// @Failure        401 {object} errs.HTTPError
// @Router         /resolve [post]
func (rc *RuntimeController) Resolve(
	ctx *gin.Context,
) {
	appId, err := appIdFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	body := new(request.ResolveRequest)

	if err := ctx.ShouldBindJSON(body); err != nil {
		ctx.Error(
			&errs.BindingError{
				Err:  err,
				Zone: errs.Body,
				Code: http.StatusBadRequest,
			},
		)
		return
	}

	result, err := rc.domain.Resolve(
		ctx.Request.Context(),
		body.ToDomain(appId),
	)
	if err != nil {
		ctx.Error(err)
		return
	}

	if result == nil {
		ctx.Status(http.StatusNoContent)
		return
	}

	ctx.JSON(
		http.StatusOK,
		response.NewResolveResult(result),
	)
}

// UpdateApp godoc
//
// @Summary        Update an application
// @Tags           Apps
// @Accept         json
// @Produce        json
// @Param          appId path string true "App ID" format(uuid)
// @Param          body body request.UpdateAppRequest true "Fields to update"
// @Success        200 {object} response.App
// @Failure        400 {object} errs.HTTPError
// @Failure        404 {object} errs.HTTPError
// @Router         /apps/{appId} [patch]
func (rc *RuntimeController) UpdateApp(ctx *gin.Context) {
	appId, err := pathUUID(ctx, "appId")
	if err != nil {
		ctx.Error(err)
		return
	}

	body := new(request.UpdateAppRequest)

	if err := ctx.ShouldBindJSON(body); err != nil {
		ctx.Error(&errs.BindingError{Err: err, Zone: errs.Body, Code: http.StatusBadRequest})
		return
	}

	updated, err := rc.domain.UpdateApp(
		ctx.Request.Context(),
		appId,
		body.Name,
		body.AllowedOrigins,
	)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, response.NewApp(updated))
}

func appIdFromContext(
	ctx *gin.Context,
) (uuid.UUID, error) {
	value, ok := ctx.Get("app_id")

	if !ok {
		return uuid.Nil, &errs.PermissionError{
			Err:  errMissingAppKey,
			Code: http.StatusUnauthorized,
		}
	}

	appId, ok := value.(uuid.UUID)

	if !ok {
		return uuid.Nil, &errs.PermissionError{
			Err:  errMissingAppKey,
			Code: http.StatusUnauthorized,
		}
	}

	return appId, nil
}
