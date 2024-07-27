package service

import (
	"context"

	"github.com/felixlambertv/go-cleanplate/internal/controller/request"
	"github.com/felixlambertv/go-cleanplate/internal/controller/response"
	"github.com/felixlambertv/go-cleanplate/internal/model"
	"github.com/felixlambertv/go-cleanplate/pkg/consttype"
	"github.com/felixlambertv/go-cleanplate/pkg/utils"
	"gorm.io/gorm"
)

// IUserService Interface
type (
	IUserService interface {
		WithTrx(trxHandle *gorm.DB) IUserService
		CreateUser(req request.CreateUserRequest) (*response.UserResponse, error)
		UpdateUserCountry(req request.UpdateUserCountryRequest, userID uint) (*response.UserResponse, error)
		GetUser(id uint) (*response.UserResponse, error)
		GetUsers(paginationReq model.Pagination) (*model.Pagination, error)
		DeleteUser(id uint) error
	}

	IAuthService interface {
		Login(req request.LoginRequest) (*response.UserResponse, *utils.TokenHeader, error)
		Register(req request.RegisterRequest) (*response.UserResponse, *utils.TokenHeader, error)
		ForgotPassword(req request.ForgotPasswordRequest) error
		ResetPassword(req request.ResetPasswordRequest) error
		SendVerificationEmail(id uint, token int) error
		VerifyToken(req request.VerifyTokenRequest) error
		SendResetPasswordEmail(id uint, token string) error
		RefreshAuthToken(token string) (*response.UserResponse, *utils.TokenHeader, error)
	}

	IMailService interface {
		SendEmail(emailData request.SendEmailRequest) error
	}

	IQueueService interface {
		SendMessage(messageBody string, messageType consttype.QueueType) error
		ReceiveMessage() error
	}

	IMediaService interface {
		UploadMedia(req request.MediaUploadRequest, ctx context.Context) (string, error)
	}
)
