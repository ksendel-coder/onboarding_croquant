package server

import (
	"github.com/Vanady39/cluer/onboarding/internal/controllers"
	"github.com/Vanady39/cluer/platform/config"
	"github.com/Vanady39/cluer/platform/middlewares"
	"github.com/Vanady39/cluer/platform/serve"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

type CreateStruct struct {
	Logger            *zerolog.Logger
	RuntimeController controllers.RuntimeControllerInterface
}

func NewServer(cfg *config.ServerConfig, createStruct *CreateStruct) *serve.Server {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	router.Use(
		gin.Logger(),
		gin.Recovery(),
		middlewares.RuntimeCORS(),
		middlewares.ErrorHandler(createStruct.Logger),
	)

	AddDocsForDebugVersion(router)

	v1 := router.Group("/v1")
	{
		v1.GET(
			"/health",
			createStruct.RuntimeController.Health,
		)

		v1.POST(
			"/apps",
			createStruct.RuntimeController.CreateApp,
		)

		v1.GET(
			"/apps",
			createStruct.RuntimeController.ListApps,
		)

		v1.PATCH(
			"/apps/:appId",
			createStruct.RuntimeController.UpdateApp,
		)
	}

	createStruct.Logger.Debug().Msg("Routes initialized")

	return serve.New(cfg, createStruct.Logger, router.Handler())
}
