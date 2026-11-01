package controllers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ksendel-coder/onboarding_croquant/onboarding/internal/domains"
	"github.com/ksendel-coder/onboarding_croquant/onboarding/internal/http/request"
	"github.com/ksendel-coder/onboarding_croquant/platform/errs"
)

type EventDomainInterface interface {
	Record(context.Context, domains.TourEvent) (bool, error)
	Analytics(context.Context, uuid.UUID, uuid.UUID) (*domains.TourAnalytics, error)
}

type EventController struct {
	domain EventDomainInterface
}

func NewEventController(domain EventDomainInterface) *EventController {
	return &EventController{domain: domain}
}

func (ec *EventController) Record(ctx *gin.Context) {
	appId, err := appIdFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	var body request.RecordEventRequest
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.Error(&errs.BindingError{Err: err, Zone: errs.Body, Code: http.StatusBadRequest})
		return
	}

	created, err := ec.domain.Record(ctx.Request.Context(), body.ToDomain(appId))
	if err != nil {
		ctx.Error(err)
		return
	}
	if created {
		ctx.Status(http.StatusAccepted)
		return
	}

	ctx.Status(http.StatusOK)
}

func (ec *EventController) Analytics(ctx *gin.Context) {
	tourId, err := pathUUID(ctx, "tourId")
	if err != nil {
		ctx.Error(err)
		return
	}

	versionId, err := uuid.Parse(ctx.Query("version_id"))
	if err != nil {
		ctx.Error(&errs.BindingError{Err: err, Zone: errs.Query, Code: http.StatusBadRequest})
		return
	}

	result, err := ec.domain.Analytics(ctx.Request.Context(), tourId, versionId)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, result)
}
