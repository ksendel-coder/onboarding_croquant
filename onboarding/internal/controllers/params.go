package controllers

import (
	"net/http"

	"github.com/ksendel-coder/onboarding_croquant/platform/errs"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func pathUUID(ctx *gin.Context, name string) (uuid.UUID, error) {
	id, err := uuid.Parse(ctx.Param(name))
	if err != nil {
		return uuid.Nil, &errs.BindingError{Err: err, Zone: errs.URI, Code: http.StatusBadRequest}
	}
	return id, nil
}
