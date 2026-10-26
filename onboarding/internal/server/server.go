package server

import (
	"github.com/gin-gonic/gin"
	"github.com/ksendel-coder/onboarding_croquant/onboarding/internal/controllers"
	"github.com/ksendel-coder/onboarding_croquant/onboarding/internal/domains"
	appmw "github.com/ksendel-coder/onboarding_croquant/onboarding/internal/middlewares"
	"github.com/ksendel-coder/onboarding_croquant/platform/config"
	"github.com/ksendel-coder/onboarding_croquant/platform/middlewares"
	"github.com/ksendel-coder/onboarding_croquant/platform/serve"
	"github.com/rs/zerolog"
)

type CreateStruct struct {
	Logger            *zerolog.Logger
	RuntimeDomain     domains.RuntimeDomainInterface
	RuntimeController controllers.RuntimeControllerInterface
	TourController    controllers.TourControllerInterface
	HintController    controllers.HintControllerInterface
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
		v1.GET("/health", createStruct.RuntimeController.Health)
		v1.POST("/apps", createStruct.RuntimeController.CreateApp)
		v1.GET("/apps", createStruct.RuntimeController.ListApps)
		v1.PATCH("/apps/:appId", createStruct.RuntimeController.UpdateApp)

		runtime := v1.Group("")
		{
			runtime.Use(
				appmw.AppKeyAuth(
					createStruct.RuntimeDomain,
				),
			)

			runtime.POST(
				"/resolve",
				createStruct.RuntimeController.Resolve,
			)
		}

		tours := v1.Group("/tours")
		{
			tours.POST("", createStruct.TourController.Create)
			tours.GET("", createStruct.TourController.List)

			tourID := tours.Group("/:tourId")
			{
				tourID.PATCH("", createStruct.TourController.UpdateMeta)
				tourID.DELETE("", createStruct.TourController.Archive)

				tourID.POST("/hints", createStruct.HintController.Create)
				tourID.GET("/hints", createStruct.HintController.List)
				tourID.PATCH("/hints/:hintId", createStruct.HintController.Update)
				tourID.DELETE("/hints/:hintId", createStruct.HintController.Delete)
				tourID.PUT("/hints/order", createStruct.HintController.Reorder)
			}
		}
	}

	createStruct.Logger.Debug().Msg("Routes initialized")

	return serve.New(cfg, createStruct.Logger, router.Handler())
}
