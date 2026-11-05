package middlewares

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func RuntimeCORS() gin.HandlerFunc {
	return func(c *gin.Context) {

		if origin := c.GetHeader("Origin"); origin != "" {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Vary", "Origin")
		}

		c.Header(
			"Access-Control-Allow-Methods",
			"GET, POST, PATCH, PUT, DELETE, OPTIONS",
		)

		c.Header(
			"Access-Control-Allow-Headers",
			"Content-Type, X-App-Key, Authorization",
		)

		c.Header(
			"Access-Control-Expose-Headers",
			"Location",
		)

		c.Header(
			"Access-Control-Max-Age",
			"600",
		)

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
