package server

import (
	"github.com/gin-gonic/gin"
	"github.com/ksendel-coder/onboarding_croquant/onboarding/internal/controllers"
	"github.com/ksendel-coder/onboarding_croquant/platform/config"
	"github.com/ksendel-coder/onboarding_croquant/platform/middlewares"
	"github.com/ksendel-coder/onboarding_croquant/platform/serve"
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
