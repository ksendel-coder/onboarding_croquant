package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ksendel-coder/onboarding_croquant/onboarding/internal/domains"
	"github.com/ksendel-coder/onboarding_croquant/onboarding/internal/middlewares"
	"github.com/ksendel-coder/onboarding_croquant/platform/errs"
)

func currentUserId(
	ctx *gin.Context,
) (uuid.UUID, error) {
	value, ok := ctx.Get(
		middlewares.ContextUserId,
	)
	if !ok {
		return uuid.Nil,
			&errs.PermissionError{
				Err:  domains.ErrUnauthorized,
				Code: http.StatusUnauthorized,
			}
	}

	userId, ok := value.(uuid.UUID)
	if !ok {
		return uuid.Nil,
			&errs.PermissionError{
				Err:  domains.ErrUnauthorized,
				Code: http.StatusUnauthorized,
			}
	}

	return userId, nil
}

func appIdFromContext(
	ctx *gin.Context,
) (uuid.UUID, error) {
	value, ok := ctx.Get(
		middlewares.ContextAppId,
	)
	if !ok {
		return uuid.Nil,
			&errs.PermissionError{
				Err:  domains.ErrUnauthorized,
				Code: http.StatusUnauthorized,
			}
	}

	appId, ok := value.(uuid.UUID)
	if !ok {
		return uuid.Nil,
			&errs.PermissionError{
				Err:  domains.ErrUnauthorized,
				Code: http.StatusUnauthorized,
			}
	}

	return appId, nil
}
