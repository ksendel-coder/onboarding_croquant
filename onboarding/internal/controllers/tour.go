package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ksendel-coder/onboarding_croquant/onboarding/internal/domains"
	"github.com/ksendel-coder/onboarding_croquant/onboarding/internal/http/request"
	"github.com/ksendel-coder/onboarding_croquant/onboarding/internal/http/response"
	"github.com/ksendel-coder/onboarding_croquant/platform/errs"
)

type (
	TourController struct {
		domain domains.TourDomainInterface
	}

	TourControllerInterface interface {
		Create(ctx *gin.Context)
		List(ctx *gin.Context)
		UpdateMeta(ctx *gin.Context)
		Archive(ctx *gin.Context)
		Publish(ctx *gin.Context)
		UpdateSettings(ctx *gin.Context)
		ListVersions(ctx *gin.Context)
		RestoreVersion(ctx *gin.Context)
	}
)

func NewTourController(domain domains.TourDomainInterface) *TourController {
	return &TourController{domain: domain}
}

func (tc *TourController) Create(ctx *gin.Context) {
	appId, err := requiredUUIDQuery(ctx, "appId")
	if err != nil {
		ctx.Error(err)
		return
	}

	body := new(request.Tour)
	if err := ctx.ShouldBindJSON(body); err != nil {
		ctx.Error(&errs.BindingError{Err: err, Zone: errs.Body, Code: http.StatusBadRequest})
		return
	}

	created, err := tc.domain.Create(ctx.Request.Context(), appId, body.ToDomain())
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.Header("Location", "/v1/tours/"+created.Id.String())
	ctx.JSON(http.StatusCreated, response.NewTour(created))
}

func (tc *TourController) List(ctx *gin.Context) {
	appId, err := requiredUUIDQuery(ctx, "appId")
	if err != nil {
		ctx.Error(err)
		return
	}

	tours, err := tc.domain.List(ctx.Request.Context(), appId)
	if err != nil {
		ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, response.NewTours(tours))
}

func (tc *TourController) UpdateMeta(ctx *gin.Context) {
	tourId, err := pathUUID(ctx, "tourId")
	if err != nil {
		ctx.Error(err)
		return
	}

	patch := new(request.TourMetaPatch)
	if err := ctx.ShouldBindJSON(patch); err != nil {
		ctx.Error(&errs.BindingError{Err: err, Zone: errs.Body, Code: http.StatusBadRequest})
		return
	}

	updated, err := tc.domain.UpdateMeta(ctx.Request.Context(), tourId, patch.ToDomain())
	if err != nil {
		ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, response.NewTour(updated))
}

func (tc *TourController) Archive(ctx *gin.Context) {
	tourId, err := pathUUID(ctx, "tourId")
	if err != nil {
		ctx.Error(err)
		return
	}

	if err := tc.domain.Archive(ctx.Request.Context(), tourId); err != nil {
		ctx.Error(err)
		return
	}
	ctx.Status(http.StatusNoContent)
}

func (tc *TourController) Publish(ctx *gin.Context) {
	tourId, err := pathUUID(ctx, "tourId")
	if err != nil {
		ctx.Error(err)
		return
	}
	result, err := tc.domain.Publish(ctx.Request.Context(), tourId)
	if err != nil {
		ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, result)
}

func (tc *TourController) UpdateSettings(ctx *gin.Context) {
	tourId, err := pathUUID(ctx, "tourId")
	if err != nil {
		ctx.Error(err)
		return
	}

	var body request.TourSettingsPatch
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.Error(&errs.BindingError{Err: err, Zone: errs.Body, Code: http.StatusBadRequest})
		return
	}

	if err := tc.domain.UpdateSettings(ctx.Request.Context(), tourId, body.ToDomain()); err != nil {
		ctx.Error(err)
		return
	}

	ctx.Status(http.StatusNoContent)
}

func (tc *TourController) ListVersions(ctx *gin.Context) {
	tourId, err := pathUUID(ctx, "tourId")
	if err != nil {
		ctx.Error(err)
		return
	}

	versions, err := tc.domain.ListVersions(ctx.Request.Context(), tourId)
	if err != nil {
		ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, versions)
}

func (tc *TourController) RestoreVersion(ctx *gin.Context) {
	tourId, err := pathUUID(ctx, "tourId")
	if err != nil {
		ctx.Error(err)
		return
	}
	sourceVersionId, err := pathUUID(ctx, "versionId")
	if err != nil {
		ctx.Error(err)
		return
	}

	var body request.RestoreVersionRequest
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.Error(&errs.BindingError{
			Err: err, Zone: errs.Body, Code: http.StatusBadRequest,
		})
		return
	}
	if !body.ConfirmReplaceDraft {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "confirm_replace_draft must be true",
		})
		return
	}

	result, err := tc.domain.RestoreVersion(ctx.Request.Context(), tourId, sourceVersionId)
	if err != nil {
		ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, result)
}
