package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ksendel-coder/onboarding_croquant/onboarding/internal/domains"
	"github.com/ksendel-coder/onboarding_croquant/onboarding/internal/http/request"
	"github.com/ksendel-coder/onboarding_croquant/onboarding/internal/http/response"
	"github.com/ksendel-coder/onboarding_croquant/onboarding/internal/middlewares"
	"github.com/ksendel-coder/onboarding_croquant/platform/errs"
)

type (
	AuthController struct {
		domain domains.AuthDomainInterface
	}

	AuthControllerInterface interface {
		Register(ctx *gin.Context)
		Login(ctx *gin.Context)
		Me(ctx *gin.Context)
		Logout(ctx *gin.Context)
	}
)

func NewAuthController(
	domain domains.AuthDomainInterface,
) *AuthController {
	return &AuthController{
		domain: domain,
	}
}

func (ac *AuthController) Register(
	ctx *gin.Context,
) {
	body := new(request.RegisterRequest)

	if err := ctx.ShouldBindJSON(
		body,
	); err != nil {
		ctx.Error(
			&errs.BindingError{
				Err:  err,
				Zone: errs.Body,
				Code: http.StatusBadRequest,
			},
		)
		return
	}

	user, token, err := ac.domain.Register(
		ctx.Request.Context(),
		body.Email,
		body.Password,
	)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(
		http.StatusCreated,
		response.NewAuth(
			user,
			token,
		),
	)
}

func (ac *AuthController) Login(
	ctx *gin.Context,
) {
	body := new(request.LoginRequest)

	if err := ctx.ShouldBindJSON(
		body,
	); err != nil {
		ctx.Error(
			&errs.BindingError{
				Err:  err,
				Zone: errs.Body,
				Code: http.StatusBadRequest,
			},
		)
		return
	}

	user, token, err := ac.domain.Login(
		ctx.Request.Context(),
		body.Email,
		body.Password,
	)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(
		http.StatusOK,
		response.NewAuth(
			user,
			token,
		),
	)
}

func (ac *AuthController) Me(
	ctx *gin.Context,
) {
	value, ok := ctx.Get(
		middlewares.ContextUser,
	)
	if !ok {
		ctx.Error(
			&errs.PermissionError{
				Err:  domains.ErrUnauthorized,
				Code: http.StatusUnauthorized,
			},
		)
		return
	}

	user, ok := value.(*domains.User)
	if !ok {
		ctx.Error(
			&errs.PermissionError{
				Err:  domains.ErrUnauthorized,
				Code: http.StatusUnauthorized,
			},
		)
		return
	}

	ctx.JSON(
		http.StatusOK,
		response.NewUser(user),
	)
}

func (ac *AuthController) Logout(
	ctx *gin.Context,
) {
	value, ok := ctx.Get(
		middlewares.ContextSessionToken,
	)
	if !ok {
		ctx.Error(
			&errs.PermissionError{
				Err:  domains.ErrUnauthorized,
				Code: http.StatusUnauthorized,
			},
		)
		return
	}

	token, ok := value.(string)
	if !ok {
		ctx.Error(
			&errs.PermissionError{
				Err:  domains.ErrUnauthorized,
				Code: http.StatusUnauthorized,
			},
		)
		return
	}

	if err := ac.domain.Logout(
		ctx.Request.Context(),
		token,
	); err != nil {
		ctx.Error(err)
		return
	}

	ctx.Status(
		http.StatusNoContent,
	)
}
