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

func requiredUUIDQuery(
	ctx *gin.Context,
	name string,
) (uuid.UUID, error) {
	raw := ctx.Query(name)

	if raw == "" {
		return uuid.Nil,
			&errs.BindingError{
				Err:  errMissingAppId,
				Zone: errs.Query,
				Code: http.StatusBadRequest,
			}
	}

	id, err := uuid.Parse(raw)

	if err != nil {
		return uuid.Nil,
			&errs.BindingError{
				Err:  err,
				Zone: errs.Query,
				Code: http.StatusBadRequest,
			}
	}

	return id, nil
}
