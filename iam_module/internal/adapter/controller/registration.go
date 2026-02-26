package controller

import (
	"base-be-golang/shared/base"
	"base-be-golang/shared/payload"
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rdhmuhammad/base-be-golang/iam-module/internal/core/constant"
	"github.com/rdhmuhammad/base-be-golang/iam-module/internal/core/usecase/registration"
	"gorm.io/gorm"
)

type AuthController struct {
	base.BaseController
	uc AuthUsecaseInterface
}

func NewAuthController(db *gorm.DB, ctrl base.BaseController) AuthController {
	return AuthController{
		BaseController: ctrl,
		uc:             registration.NewUsecase(db),
	}
}

type AuthUsecaseInterface interface {
	Register(ctx context.Context, request registration.RegisterRequest) (registration.RegisterResponse, error)
	Logout(ctx context.Context, role string) error
	Login(ctx context.Context, request registration.LoginRequest) (registration.LoginResponse, error)
	VerifyAcc(ctx context.Context, request registration.VerifyAccRequest) (registration.VerifyAccResponse, error)
	ResendOTP(ctx context.Context, request registration.SendOtpRequest) error
}

func (ctrl AuthController) Logout(c *gin.Context, role string) {
	err := ctrl.uc.Logout(c.Request.Context(), role)
	ctrl.Mapper.NewResponse(c, payload.NewSuccessResponseNoData(constant.LogoutSuccess.String()), err)
}

func (ctrl AuthController) Login(c *gin.Context, role string) {
	var request registration.LoginRequest
	if errs := ctrl.Enigma.BindAndValidate(c, &request); len(errs) > 0 {
		c.JSON(http.StatusBadRequest, payload.DefaultInvalidInputFormResponse(errs))
		return
	}
	request.Role = role
	result, err := ctrl.uc.Login(c.Request.Context(), request)
	ctrl.Mapper.NewResponse(c, payload.NewSuccessResponse(result, constant.LoginSuccess.String()), err)
}

func (ctrl AuthController) Register(c *gin.Context) {
	var request registration.RegisterRequest
	if errs := ctrl.Enigma.BindAndValidate(c, &request); len(errs) > 0 {
		c.JSON(http.StatusBadRequest, payload.DefaultInvalidInputFormResponse(errs))
		return
	}

	result, err := ctrl.uc.Register(c.Request.Context(), request)
	ctrl.Mapper.NewResponse(c, payload.NewSuccessResponse(result, constant.RegisterSuccess.String()), err)
}

func (ctrl AuthController) VerifyAcc(c *gin.Context) {
	var request registration.VerifyAccRequest
	if errs := ctrl.Enigma.BindAndValidate(c, &request); len(errs) > 0 {
		c.JSON(http.StatusBadRequest, payload.DefaultInvalidInputFormResponse(errs))
		return
	}

	result, err := ctrl.uc.VerifyAcc(c.Request.Context(), request)
	ctrl.Mapper.NewResponse(c, payload.NewSuccessResponse(result, constant.VerifyOtpSuccess.String()), err)
}

func (ctrl AuthController) ResendOTP(c *gin.Context) {
	var request registration.SendOtpRequest
	if errs := ctrl.Enigma.BindAndValidate(c, &request); len(errs) > 0 {
		c.JSON(http.StatusBadRequest, payload.DefaultInvalidInputFormResponse(errs))
	}

	err := ctrl.uc.ResendOTP(c.Request.Context(), request)
	ctrl.Mapper.NewResponse(c, payload.NewSuccessResponseNoData(constant.ResendOtpSuccess.String()), err)
}

func (ctrl AuthController) Route(router *gin.RouterGroup) {
	userAuth := router.Group("/auth")
	userAuth.POST("/register",
		ctrl.Idem.Idempotent(
			"/register",
			"username",
			time.Millisecond*2,
		),
		ctrl.Register,
	)

	userAuth.POST("/login",
		func(c *gin.Context) {
			ctrl.Login(c, constant.ContextMobile)
		},
	)

	userAuth.POST("/login/admin",
		func(c *gin.Context) {
			ctrl.Login(c, constant.ContextDashboard)
		},
	)

	userAuth.POST("/logout",
		ctrl.Security.Validate(),
		ctrl.Security.Authorize(constant.RoleIsAdmin, constant.RoleIsUser),
		func(c *gin.Context) {
			ctrl.Logout(c, constant.ContextMobile)
		})

	userAuth.POST("/logout/admin",
		ctrl.Security.Validate(),
		ctrl.Security.Authorize(constant.RolesIsMobile),
		func(c *gin.Context) {
			ctrl.Logout(c, constant.ContextDashboard)
		})

	userAuth.POST(
		"/verify-acc",
		ctrl.VerifyAcc,
	)

	userAuth.POST(
		"/resend-otp",
		ctrl.Idem.Idempotent(
			"/resend-otp",
			"username",
			time.Minute*1,
		),
		ctrl.ResendOTP,
	)
}
