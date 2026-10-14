package middlewares

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ksendel-coder/onboarding_croquant/onboarding/internal/domains"
	platform "github.com/ksendel-coder/onboarding_croquant/platform/middlewares"
)

const (
	ContextApp   = "app"
	ContextAppId = "app_id"
)

func AppKeyAuth(
	runtime domains.RuntimeDomainInterface,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.GetHeader("X-App-Key")

		if key == "" {
			c.Error(
				platform.NewAuthHeaderError(
					http.StatusUnauthorized,
					"X-App-Key header is required",
					domains.ErrAppKeyRequired,
				),
			)

			c.Abort()
			return
		}

		app, err := runtime.AppByKey(
			c.Request.Context(),
			key,
		)

		if err != nil {
			c.Error(
				platform.NewAuthHeaderError(
					http.StatusUnauthorized,
					"Unknown application key",
					domains.ErrAppNotFound,
				),
			)

			c.Abort()
			return
		}

		origin := c.GetHeader("Origin")

		if origin != "" {
			if !domains.OriginAllowed(
				app.AllowedOrigins,
				origin,
			) {
				c.Error(
					platform.NewAuthHeaderError(
						http.StatusForbidden,
						"Origin is not allowed for this application",
						domains.ErrOriginNotAllowed,
					),
				)

				c.Abort()
				return
			}

			c.Header(
				"Access-Control-Allow-Origin",
				origin,
			)
			c.Header("Vary", "Origin")
		}

		c.Set(ContextApp, app)
		c.Set(ContextAppId, app.Id)

		c.Next()
	}
}
