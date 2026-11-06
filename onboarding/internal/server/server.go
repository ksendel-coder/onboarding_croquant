package server

import (
	"github.com/gin-gonic/gin"
	"github.com/ksendel-coder/onboarding_croquant/onboarding/internal/controllers"
	"github.com/ksendel-coder/onboarding_croquant/onboarding/internal/domains"
	"github.com/ksendel-coder/onboarding_croquant/onboarding/internal/middlewares"
	"github.com/ksendel-coder/onboarding_croquant/platform/config"
	platformmw "github.com/ksendel-coder/onboarding_croquant/platform/middlewares"
	"github.com/ksendel-coder/onboarding_croquant/platform/serve"
	"github.com/rs/zerolog"
)

type CreateStruct struct {
	Logger            *zerolog.Logger
	RuntimeDomain     domains.RuntimeDomainInterface
	RuntimeController controllers.RuntimeControllerInterface
	TourController    controllers.TourControllerInterface
	HintController    controllers.HintControllerInterface
	AuthDomain        domains.AuthDomainInterface
	AuthController    controllers.AuthControllerInterface
	EventController   *controllers.EventController
}

func NewServer(cfg *config.ServerConfig, createStruct *CreateStruct) *serve.Server {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	router.Use(
		gin.Logger(),
		gin.Recovery(),
		platformmw.RuntimeCORS(),
		platformmw.ErrorHandler(createStruct.Logger),
	)

	v1 := router.Group("/v1")
	{
		v1.GET("/health", createStruct.RuntimeController.Health)
		apps := v1.Group("/apps")
		apps.Use(
			middlewares.AdminAuth(
				createStruct.AuthDomain,
			),
		)
		{
			apps.POST(
				"",
				createStruct.RuntimeController.CreateApp,
			)

			apps.GET(
				"",
				createStruct.RuntimeController.ListApps,
			)

			apps.PATCH(
				"/:appId",
				createStruct.RuntimeController.UpdateApp,
			)
		}

		runtime := v1.Group("")
		{
			runtime.Use(
				middlewares.AppKeyAuth(
					createStruct.RuntimeDomain,
				),
			)

			runtime.POST(
				"/resolve",
				createStruct.RuntimeController.Resolve,
			)
			runtime.POST("/events", createStruct.EventController.Record)
		}

		tours := v1.Group("/tours")
		tours.Use(
			middlewares.AdminAuth(
				createStruct.AuthDomain,
			),
		)
		{
			tours.POST(
				"",
				middlewares.RequireAppAccess(
					createStruct.AuthDomain,
				),
				createStruct.TourController.Create,
			)

			tours.GET(
				"",
				middlewares.RequireAppAccess(
					createStruct.AuthDomain,
				),
				createStruct.TourController.List,
			)

			tourID := tours.Group("/:tourId")
			tourID.Use(
				middlewares.RequireTourAccess(
					createStruct.AuthDomain,
				),
			)
			{
				tourID.PATCH(
					"/settings",
					createStruct.TourController.UpdateSettings,
				)

				tourID.PATCH(
					"",
					createStruct.TourController.UpdateMeta,
				)

				tourID.DELETE(
					"",
					createStruct.TourController.Archive,
				)

				tourID.POST(
					"/hints",
					createStruct.HintController.Create,
				)

				tourID.GET(
					"/hints",
					createStruct.HintController.List,
				)

				tourID.PATCH(
					"/hints/:hintId",
					createStruct.HintController.Update,
				)

				tourID.DELETE(
					"/hints/:hintId",
					createStruct.HintController.Delete,
				)

				tourID.PUT(
					"/hints/order",
					createStruct.HintController.Reorder,
				)

				tourID.POST(
					"/publish",
					createStruct.TourController.Publish,
				)
			}
		}
	}

	auth := v1.Group("/auth")
	{
		auth.POST(
			"/register",
			createStruct.AuthController.Register,
		)

		auth.POST(
			"/login",
			createStruct.AuthController.Login,
		)

		protectedAuth := auth.Group("")
		protectedAuth.Use(
			middlewares.AdminAuth(
				createStruct.AuthDomain,
			),
		)
		{
			protectedAuth.GET(
				"/me",
				createStruct.AuthController.Me,
			)

			protectedAuth.POST(
				"/logout",
				createStruct.AuthController.Logout,
			)
		}
	}

	createStruct.Logger.Debug().Msg("Routes initialized")

	return serve.New(cfg, createStruct.Logger, router.Handler())
}
