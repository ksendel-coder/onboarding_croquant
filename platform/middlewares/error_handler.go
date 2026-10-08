package middlewares

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ksendel-coder/onboarding_croquant/platform/errs"
	"github.com/rs/zerolog"
)

func ErrorHandler(logger *zerolog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		// Check if errors occurred during request handling
		if len(c.Errors) > 0 {
			logger.Trace().Int("err_count", len(c.Errors)).Msg("One or more errors occurred")

			// Check and prepare the first error that occurred
			// If we have more than one or one of a different type, something is probably broken
			switch err := c.Errors.Last().Err.(type) {
			case errs.BaseError:
				c.AbortWithStatusJSON(err.StatusCode(), err.ToHTTPError())
			default:
				logger.Error().Err(err).Msg("Error handler got an error of unknown type")
				c.AbortWithStatusJSON(http.StatusInternalServerError,
					errs.HTTPError{
						Error:   "failed to process errors",
						Message: "An unknown error has occurred",
					},
				)
				return
			}
		}
	}
}
