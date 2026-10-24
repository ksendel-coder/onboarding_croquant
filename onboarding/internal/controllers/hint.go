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
	HintController struct {
		domain domains.HintDomainInterface
	}

	HintControllerInterface interface {
		Create(ctx *gin.Context)
		List(ctx *gin.Context)
		Update(ctx *gin.Context)
		Delete(ctx *gin.Context)
		Reorder(ctx *gin.Context)
	}
)

func NewHintController(
	domain domains.HintDomainInterface,
) *HintController {
	return &HintController{
		domain: domain,
	}
}

func (hc *HintController) Create(
	ctx *gin.Context,
) {
	tourId, err := pathUUID(
		ctx,
		"tourId",
	)
	if err != nil {
		ctx.Error(err)
		return
	}

	body := new(request.Hint)

	if err := ctx.ShouldBindJSON(
		body,
	); err != nil {
		ctx.Error(
			&errs.BindingError{
				Err:  err,
				Zone: errs.Body,
				Code: http.StatusBadRequest,
			},
		)
		return
	}

	created, err := hc.domain.Create(
		ctx.Request.Context(),
		tourId,
		body.ToDomain(),
	)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.Header(
		"Location",
		"/v1/tours/"+
			tourId.String()+
			"/hints/"+
			created.Id.String(),
	)

	ctx.JSON(
		http.StatusCreated,
		response.NewHint(created),
	)
}

func (hc *HintController) List(
	ctx *gin.Context,
) {
	tourId, err := pathUUID(
		ctx,
		"tourId",
	)
	if err != nil {
		ctx.Error(err)
		return
	}

	hints, err := hc.domain.ListByTour(
		ctx.Request.Context(),
		tourId,
	)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(
		http.StatusOK,
		response.NewHints(hints),
	)
}

func (hc *HintController) Update(
	ctx *gin.Context,
) {
	tourId, err := pathUUID(
		ctx,
		"tourId",
	)
	if err != nil {
		ctx.Error(err)
		return
	}

	hintId, err := pathUUID(
		ctx,
		"hintId",
	)
	if err != nil {
		ctx.Error(err)
		return
	}

	body := new(request.HintPatch)

	if err := ctx.ShouldBindJSON(
		body,
	); err != nil {
		ctx.Error(
			&errs.BindingError{
				Err:  err,
				Zone: errs.Body,
				Code: http.StatusBadRequest,
			},
		)
		return
	}

	updated, err := hc.domain.Update(
		ctx.Request.Context(),
		tourId,
		hintId,
		body.ToDomain(),
	)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(
		http.StatusOK,
		response.NewHint(updated),
	)
}

func (hc *HintController) Delete(
	ctx *gin.Context,
) {
	tourId, err := pathUUID(
		ctx,
		"tourId",
	)
	if err != nil {
		ctx.Error(err)
		return
	}

	hintId, err := pathUUID(
		ctx,
		"hintId",
	)
	if err != nil {
		ctx.Error(err)
		return
	}

	if err := hc.domain.Delete(
		ctx.Request.Context(),
		tourId,
		hintId,
	); err != nil {
		ctx.Error(err)
		return
	}

	ctx.Status(
		http.StatusNoContent,
	)
}

func (hc *HintController) Reorder(
	ctx *gin.Context,
) {
	tourId, err := pathUUID(
		ctx,
		"tourId",
	)
	if err != nil {
		ctx.Error(err)
		return
	}

	body := new(request.ReorderRequest)

	if err := ctx.ShouldBindJSON(
		body,
	); err != nil {
		ctx.Error(
			&errs.BindingError{
				Err:  err,
				Zone: errs.Body,
				Code: http.StatusBadRequest,
			},
		)
		return
	}

	hints, err := hc.domain.Reorder(
		ctx.Request.Context(),
		tourId,
		body.HintIds,
	)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(
		http.StatusOK,
		response.NewHints(hints),
	)
}
