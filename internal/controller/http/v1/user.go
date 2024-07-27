package v1

import (
	"net/http"
	"strings"

	"github.com/felixlambertv/go-cleanplate/config"
	"github.com/felixlambertv/go-cleanplate/internal/controller/request"
	"github.com/felixlambertv/go-cleanplate/internal/controller/response"
	"github.com/felixlambertv/go-cleanplate/internal/middleware"
	"github.com/felixlambertv/go-cleanplate/internal/model"
	"github.com/felixlambertv/go-cleanplate/internal/service"
	"github.com/felixlambertv/go-cleanplate/pkg/consttype"
	"github.com/felixlambertv/go-cleanplate/pkg/logger"
	"github.com/felixlambertv/go-cleanplate/pkg/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type userRoutes struct {
	s   service.IUserService
	l   logger.Interface
	cfg *config.Config
}

func newUserRoutes(handler *gin.RouterGroup, l logger.Interface, db *gorm.DB, s service.IUserService, cfg *config.Config) {
	r := &userRoutes{l: l, s: s, cfg: cfg}

	h := handler.Group("users").Use(middleware.JWTAuthMiddleware(cfg, consttype.ADMIN))
	{
		h.GET("", r.getUser)
		h.POST("", r.createUser)
	}

	userHandler := handler.Group("users").Use(middleware.JWTAuthMiddleware(cfg, consttype.USER))
	{
		userHandler.PATCH("/country", r.updateUserCountry)
		userHandler.GET("/me", r.getCurrentUser)
		userHandler.DELETE("/delete", r.deleteUser)
	}
}

func (r *userRoutes) createUser(ctx *gin.Context) {
	var req request.CreateUserRequest

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

	user, err := r.s.CreateUser(req)
	if err != nil {
		if strings.Contains(err.Error(), "SQLSTATE 23505") {
			utils.ErrorResponse(ctx, http.StatusConflict, utils.ErrorRes{
				Message: "Duplicate email",
				Debug:   err,
				Errors:  err.Error(),
			})
		} else {
			utils.ErrorResponse(ctx, http.StatusInternalServerError, utils.ErrorRes{
				Message: "Something went wrong",
				Debug:   err,
				Errors:  err.Error(),
			})
		}
		return
	}

	utils.SuccessResponse(ctx, http.StatusOK, utils.SuccessRes{
		Message: "Success Creating new user",
		Data:    user,
	})
}

func (r *userRoutes) updateUserCountry(ctx *gin.Context) {
	var req request.UpdateUserCountryRequest

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

	user, err := r.s.UpdateUserCountry(req, loggedInUser.ID)
	if err != nil {
		utils.ErrorResponse(ctx, http.StatusInternalServerError, utils.ErrorRes{
			Message: "Something went wrong",
			Debug:   err,
			Errors:  err.Error(),
		})
		return
	}

	utils.SuccessResponse(ctx, http.StatusOK, utils.SuccessRes{
		Message: "Success updating user",
		Data:    user,
	})
}

func (r *userRoutes) getUser(ctx *gin.Context) {
	paginationReq := utils.GeneratePaginationFromRequest(ctx, model.User{})

	users, err := r.s.GetUsers(paginationReq)
	if err != nil {
		utils.ErrorResponse(ctx, http.StatusNotFound, utils.ErrorRes{
			Message: "User not found",
			Debug:   nil,
			Errors:  nil,
		})
		return
	}

	utils.SuccessResponse(ctx, http.StatusOK, utils.SuccessRes{
		Message: "Success Get Users",
		Data:    users,
	})
}

func (r *userRoutes) getCurrentUser(ctx *gin.Context) {
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

	user, err := r.s.GetUser(loggedInUser.ID)
	if err != nil {
		utils.ErrorResponse(ctx, http.StatusNotFound, utils.ErrorRes{
			Message: "User not found",
			Debug:   nil,
			Errors:  "User not found",
		})
		return
	}

	utils.SuccessResponse(ctx, http.StatusOK, utils.SuccessRes{
		Message: "Success Get User",
		Data:    user,
	})
}

func (r *userRoutes) deleteUser(ctx *gin.Context) {
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

	if loggedInUser.UserLevel == consttype.ADMIN {
		utils.ErrorResponse(ctx, http.StatusNotFound, utils.ErrorRes{
			Message: "Something went wrong",
			Debug:   nil,
			Errors:  "Admin Level can't be deleted",
		})
		return
	}

	err := r.s.DeleteUser(loggedInUser.ID)
	if err != nil {
		utils.ErrorResponse(ctx, http.StatusNotFound, utils.ErrorRes{
			Message: "Something went wrong",
			Debug:   err,
			Errors:  "Something went wrong while deleting user",
		})
		return
	}

	utils.SuccessResponse(ctx, http.StatusOK, utils.SuccessRes{
		Message: "Success Delete User",
		Data:    nil,
	})
}
