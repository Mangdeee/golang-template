package v1

import (
	"net/http"

	"github.com/felixlambertv/go-cleanplate/config"
	"github.com/felixlambertv/go-cleanplate/internal/controller/request"
	"github.com/felixlambertv/go-cleanplate/internal/controller/response"
	"github.com/felixlambertv/go-cleanplate/internal/middleware"
	"github.com/felixlambertv/go-cleanplate/internal/service"
	"github.com/felixlambertv/go-cleanplate/pkg/consttype"
	"github.com/felixlambertv/go-cleanplate/pkg/logger"
	"github.com/felixlambertv/go-cleanplate/pkg/utils"
	"github.com/gin-gonic/gin"
)

type authRoutes struct {
	l   logger.Interface
	cfg *config.Config
	s   service.IAuthService
	ms  service.IMailService
}

func newAuthRoutes(handler *gin.RouterGroup, l logger.Interface, cfg *config.Config, s service.IAuthService, ms service.IMailService) {
	r := &authRoutes{l: l, s: s, cfg: cfg, ms: ms}

	h := handler.Group("auth")
	{
		h.POST("login", r.login)
		h.POST("register", r.register)

		verifyGroup := h.Group("verify").Use(middleware.JWTAuthMiddleware(cfg, consttype.USER))
		{
			verifyGroup.POST("", r.verifyToken)
			verifyGroup.POST("send", r.sendVerifyEmail)
		}

		h.POST("forgot-password", r.forgotPassword)
		h.POST("reset-password", r.resetPassword)

		h.GET("refresh-token", r.refreshAuthToken)
	}
}

func (r *authRoutes) login(ctx *gin.Context) {
	var req request.LoginRequest

	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		ve := utils.ValidationResponse(err)

		utils.ErrorResponse(ctx, http.StatusBadRequest, utils.ErrorRes{
			Message: "request not valid",
			Debug:   nil,
			Errors:  ve,
		})
		return
	}

	user, token, err := r.s.Login(req)
	if err != nil {
		utils.ErrorResponse(ctx, http.StatusUnauthorized, utils.ErrorRes{
			Message: "Something went wrong",
			Debug:   err,
			Errors:  err.Error(),
		})
		return
	}

	res := response.AuthResponse{
		ID:                 user.ID,
		FullName:           user.FullName,
		Email:              user.Email,
		UserLevel:          user.UserLevel,
		Country:            user.Country,
		CountryCode:        user.CountryCode,
		ConfirmationSentAt: user.ConfirmationSentAt,
		ConfirmedAt:        user.ConfirmedAt,
		CreatedAt:          user.CreatedAt,
		UpdatedAt:          user.UpdatedAt,
		Token:              token.AuthToken,
		Expires:            token.AuthTokenExpires,
	}

	utils.SuccessResponse(ctx, http.StatusOK, utils.SuccessRes{
		Message: "Login Successful",
		Data:    res,
		Header:  *token,
	})
}

func (r *authRoutes) register(ctx *gin.Context) {
	var req request.RegisterRequest

	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		ve := utils.ValidationResponse(err)

		utils.ErrorResponse(ctx, http.StatusBadRequest, utils.ErrorRes{
			Message: "request not valid",
			Debug:   err,
			Errors:  ve,
		})
		return
	}

	user, token, err := r.s.Register(req)
	if err != nil {
		utils.ErrorResponse(ctx, http.StatusUnauthorized, utils.ErrorRes{
			Message: "Something went wrong while registering",
			Debug:   err,
			Errors:  err.Error(),
		})
		return
	}

	res := response.AuthResponse{
		ID:                 user.ID,
		FullName:           user.FullName,
		Email:              user.Email,
		UserLevel:          user.UserLevel,
		Country:            user.Country,
		CountryCode:        user.CountryCode,
		ConfirmationSentAt: user.ConfirmationSentAt,
		ConfirmedAt:        user.ConfirmedAt,
		CreatedAt:          user.CreatedAt,
		UpdatedAt:          user.UpdatedAt,
		Token:              token.AuthToken,
		Expires:            token.AuthTokenExpires,
	}

	utils.SuccessResponse(ctx, http.StatusOK, utils.SuccessRes{
		Message: "Register Successful",
		Data:    res,
		Header:  *token,
	})
}

func (r *authRoutes) sendVerifyEmail(ctx *gin.Context) {
	ctxUser, exists := ctx.Get("user")
	if !exists {
		utils.ErrorResponse(ctx, http.StatusNotFound, utils.ErrorRes{
			Message: "Error getting user",
			Debug:   nil,
			Errors:  "User not found",
		})
		return
	}

	loggedInUser, ok := ctxUser.(response.UserResponse)
	if !ok {
		utils.ErrorResponse(ctx, http.StatusNotFound, utils.ErrorRes{
			Message: "Error getting user",
			Debug:   nil,
			Errors:  "Unable to assert User ID",
		})
		return
	}

	token := utils.GenerateRandomToken()
	err := r.s.SendVerificationEmail(loggedInUser.ID, token)
	if err != nil {
		utils.ErrorResponse(ctx, http.StatusNotFound, utils.ErrorRes{
			Message: "Error sending verification email",
			Debug:   err,
			Errors:  err.Error(),
		})
		return
	}

	utils.SuccessResponse(ctx, http.StatusOK, utils.SuccessRes{
		Message: "Send Verification Email Successful",
		Data:    nil,
	})
}

func (r *authRoutes) verifyToken(ctx *gin.Context) {
	var req request.VerifyTokenRequest

	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		utils.ErrorResponse(ctx, http.StatusBadRequest, utils.ErrorRes{
			Message: "request not valid",
			Debug:   err,
			Errors:  err.Error(),
		})
		return
	}

	err = r.s.VerifyToken(req)
	if err != nil {
		utils.ErrorResponse(ctx, http.StatusBadRequest, utils.ErrorRes{
			Message: "Cannot verify token",
			Debug:   err,
			Errors:  err.Error(),
		})
		return
	}

	utils.SuccessResponse(ctx, http.StatusOK, utils.SuccessRes{
		Message: "Verification successful",
		Data:    nil,
	})
}

func (r *authRoutes) forgotPassword(ctx *gin.Context) {
	var req request.ForgotPasswordRequest

	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		utils.ErrorResponse(ctx, http.StatusBadRequest, utils.ErrorRes{
			Message: "request not valid",
			Debug:   err,
			Errors:  err.Error(),
		})
		return
	}

	err = r.s.ForgotPassword(req)
	if err != nil {
		utils.ErrorResponse(ctx, http.StatusBadRequest, utils.ErrorRes{
			Message: "Something went wrong",
			Debug:   err,
			Errors:  err.Error(),
		})
		return
	}

	utils.SuccessResponse(ctx, http.StatusOK, utils.SuccessRes{
		Message: "Successfully requested to forgot password",
		Data:    nil,
	})
}

func (r *authRoutes) resetPassword(ctx *gin.Context) {
	var req request.ResetPasswordRequest

	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		ve := utils.ValidationResponse(err)

		utils.ErrorResponse(ctx, http.StatusBadRequest, utils.ErrorRes{
			Message: "request not valid",
			Debug:   err,
			Errors:  ve,
		})
		return
	}

	err = r.s.ResetPassword(req)
	if err != nil {
		utils.ErrorResponse(ctx, http.StatusBadRequest, utils.ErrorRes{
			Message: "Something went wrong",
			Debug:   err,
			Errors:  err.Error(),
		})
		return
	}

	utils.SuccessResponse(ctx, http.StatusOK, utils.SuccessRes{
		Message: "Your password has been successfully changed",
		Data:    nil,
	})
}

func (r *authRoutes) refreshAuthToken(ctx *gin.Context) {
	refreshToken := ctx.Request.Header.Get("Refresh-Token")

	if refreshToken == "" {
		utils.ErrorResponse(ctx, http.StatusUnauthorized, utils.ErrorRes{
			Message: "Something went wrong",
			Debug:   nil,
			Errors:  "No Refresh token detected",
		})
		return
	}

	user, token, err := r.s.RefreshAuthToken(refreshToken)
	if err != nil {
		utils.ErrorResponse(ctx, http.StatusUnauthorized, utils.ErrorRes{
			Message: "Something went wrong",
			Debug:   err,
			Errors:  err.Error(),
		})
		return
	}

	res := response.AuthResponse{
		ID:          user.ID,
		FullName:    user.FullName,
		Email:       user.Email,
		UserLevel:   user.UserLevel,
		Country:     user.Country,
		CountryCode: user.CountryCode,
		ConfirmedAt: user.ConfirmedAt,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
		Token:       token.AuthToken,
		Expires:     token.AuthTokenExpires,
	}

	utils.SuccessResponse(ctx, http.StatusOK, utils.SuccessRes{
		Message: "Refresh Token Successful",
		Data:    res,
		Header:  *token,
	})
}
