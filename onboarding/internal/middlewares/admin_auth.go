package middlewares

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ksendel-coder/onboarding_croquant/onboarding/internal/domains"
	"github.com/ksendel-coder/onboarding_croquant/platform/errs"
	platform "github.com/ksendel-coder/onboarding_croquant/platform/middlewares"
)

const (
	ContextUser         = "admin_user"
	ContextUserId       = "admin_user_id"
	ContextSessionToken = "admin_session_token"
)

func AdminAuth(
	auth domains.AuthDomainInterface,
) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token, ok := bearerToken(
			ctx.GetHeader("Authorization"),
		)
		if !ok {
			ctx.Error(
				platform.NewAuthHeaderError(
					http.StatusUnauthorized,
					"Authorization Bearer token is required",
					domains.ErrUnauthorized,
				),
			)
			ctx.Abort()
			return
		}

		user, err := auth.Authenticate(
			ctx.Request.Context(),
			token,
		)
		if err != nil {
			ctx.Error(
				platform.NewAuthHeaderError(
					http.StatusUnauthorized,
					"Invalid or expired session",
					domains.ErrUnauthorized,
				),
			)
			ctx.Abort()
			return
		}

		ctx.Set(
			ContextUser,
			user,
		)
		ctx.Set(
			ContextUserId,
			user.Id,
		)
		ctx.Set(
			ContextSessionToken,
			token,
		)

		ctx.Next()
	}
}

func RequireAppAccess(
	auth domains.AuthDomainInterface,
) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userId, ok := contextUserId(ctx)
		if !ok {
			ctx.Error(
				&errs.PermissionError{
					Err:  domains.ErrUnauthorized,
					Code: http.StatusUnauthorized,
				},
			)
			ctx.Abort()
			return
		}

		rawAppId := ctx.Query("appId")
		if rawAppId == "" {
			ctx.Error(
				&errs.BindingError{
					Err:    errors.New("appId query parameter is required"),
					Zone:   errs.Query,
					Code:   http.StatusBadRequest,
					Detail: "appId query parameter is required",
				},
			)
			ctx.Abort()
			return
		}

		appId, err := uuid.Parse(
			rawAppId,
		)
		if err != nil {
			ctx.Error(
				&errs.BindingError{
					Err:  err,
					Zone: errs.Query,
					Code: http.StatusBadRequest,
				},
			)
			ctx.Abort()
			return
		}

		allowed, err := auth.UserHasApp(
			ctx.Request.Context(),
			userId,
			appId,
		)
		if err != nil {
			ctx.Error(err)
			ctx.Abort()
			return
		}

		if !allowed {
			ctx.Error(
				&errs.PermissionError{
					Err:  domains.ErrAccessDenied,
					Code: http.StatusForbidden,
				},
			)
			ctx.Abort()
			return
		}

		ctx.Next()
	}
}

func RequireTourAccess(
	auth domains.AuthDomainInterface,
) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userId, ok := contextUserId(ctx)
		if !ok {
			ctx.Error(
				&errs.PermissionError{
					Err:  domains.ErrUnauthorized,
					Code: http.StatusUnauthorized,
				},
			)
			ctx.Abort()
			return
		}

		tourId, err := uuid.Parse(
			ctx.Param("tourId"),
		)
		if err != nil {
			ctx.Error(
				&errs.BindingError{
					Err:  err,
					Zone: errs.URI,
					Code: http.StatusBadRequest,
				},
			)
			ctx.Abort()
			return
		}

		allowed, err := auth.UserHasTour(
			ctx.Request.Context(),
			userId,
			tourId,
		)
		if err != nil {
			ctx.Error(err)
			ctx.Abort()
			return
		}

		if !allowed {
			ctx.Error(
				&errs.PermissionError{
					Err:  domains.ErrAccessDenied,
					Code: http.StatusForbidden,
				},
			)
			ctx.Abort()
			return
		}

		ctx.Next()
	}
}

func bearerToken(
	header string,
) (string, bool) {
	const prefix = "Bearer "

	if !strings.HasPrefix(
		header,
		prefix,
	) {
		return "", false
	}

	token := strings.TrimSpace(
		strings.TrimPrefix(
			header,
			prefix,
		),
	)

	if token == "" {
		return "", false
	}

	return token, true
}

func contextUserId(
	ctx *gin.Context,
) (uuid.UUID, bool) {
	value, ok := ctx.Get(
		ContextUserId,
	)
	if !ok {
		return uuid.Nil, false
	}

	userId, ok := value.(uuid.UUID)

	return userId, ok
}
