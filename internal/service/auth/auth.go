package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/felixlambertv/go-cleanplate/config"
	"github.com/felixlambertv/go-cleanplate/internal/controller/request"
	"github.com/felixlambertv/go-cleanplate/internal/controller/response"
	"github.com/felixlambertv/go-cleanplate/internal/model"
	"github.com/felixlambertv/go-cleanplate/internal/repository"
	"github.com/felixlambertv/go-cleanplate/internal/service"
	"github.com/felixlambertv/go-cleanplate/pkg/consttype"
	"github.com/felixlambertv/go-cleanplate/pkg/utils"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthService struct {
	cfg      *config.Config
	userRepo repository.IUserRepo
	ms       service.IMailService
	qs       service.IQueueService
}

func NewAuthService(userRepo repository.IUserRepo, cfg *config.Config, ms service.IMailService, qs service.IQueueService) *AuthService {
	return &AuthService{userRepo: userRepo, cfg: cfg, ms: ms, qs: qs}
}

func (a *AuthService) Login(req request.LoginRequest) (*response.UserResponse, *utils.TokenHeader, error) {
	user, err := a.userRepo.FindByEmail(req.Email)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil, errors.New("user not found")
		}

		return nil, nil, err
	}

	err = verifyPassword(user, req.Password)
	if err != nil {
		return nil, nil, err
	}

	tokenHeader, err := a.generateAuthTokens(user)
	if err != nil {
		return nil, nil, err
	}

	return user, tokenHeader, nil
}

func (a *AuthService) Register(req request.RegisterRequest) (*response.UserResponse, *utils.TokenHeader, error) {
	var user *response.UserResponse
	req.Email = strings.ToLower(req.Email)

	user, err := a.userRepo.FindByEmail(req.Email)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, nil, errors.New("user not found")
	}

	if user != nil {
		return nil, nil, errors.New("user already exists")
	}

	hashedPassword, err := utils.EncryptPassword(req.Password)
	if err != nil {
		return nil, nil, err
	}

	userCreate := &model.User{
		FullName:  req.FullName,
		Email:     req.Email,
		Password:  hashedPassword,
		Country:   req.Country,
		UserLevel: consttype.USER,
	}

	userModel, err := a.userRepo.Store(userCreate)
	if err != nil {
		return nil, nil, err
	}

	marshaledUser, _ := json.Marshal(userModel)
	err = json.Unmarshal(marshaledUser, &user)
	if err != nil {
		return nil, nil, err
	}

	token, err := a.generateAuthTokens(user)
	if err != nil {
		return nil, nil, err
	}

	return user, token, nil
}

func (a *AuthService) ForgotPassword(req request.ForgotPasswordRequest) error {
	user, err := a.userRepo.FindByEmail(req.Email)
	if err != nil && err == gorm.ErrRecordNotFound {
		return errors.New("user not found")
	}

	token := utils.GenerateRandomStringToken(16)

	err = a.SendResetPasswordEmail(user.ID, token)
	if err != nil {
		return err
	}

	return nil
}

func (a *AuthService) ResetPassword(req request.ResetPasswordRequest) error {
	user, err := a.userRepo.FindByEmail(req.Email)
	if err != nil && err == gorm.ErrRecordNotFound {
		return errors.New("user not found")
	}

	if req.Token != user.ResetPasswordToken {
		return errors.New("token mismatch")
	}

	if !time.Now().UTC().Before(user.ResetPasswordSentAt.Add(time.Minute * 5)) {
		return errors.New("this token is expired")
	}

	hashedPassword, err := utils.EncryptPassword(req.Password)
	if err != nil {
		return err
	}

	userUpdate := model.User{
		Password:           hashedPassword,
		ResetPasswordToken: "",
	}

	_, err = a.userRepo.Update(userUpdate, user.ID)
	if err != nil {
		return err
	}

	return nil
}

func (a *AuthService) SendResetPasswordEmail(id uint, token string) error {
	user, err := a.userRepo.FindById(id)
	if err != nil {
		return err
	}

	if time.Now().UTC().Before(user.ResetPasswordSentAt.Add(time.Minute * 5)) {
		return errors.New("you already requested a reset password email in less than 5 minutes")
	}

	userUpdate := model.User{
		ResetPasswordToken:  token,
		ResetPasswordSentAt: time.Now().UTC(),
	}

	_, err = a.userRepo.Update(userUpdate, user.ID)
	if err != nil {
		return err
	}

	emailData := request.SendEmailRequest{
		Template: "reset_password.html",
		Subject:  "Reset Password",
		Name:     user.FullName,
		Email:    user.Email,
		Token:    0,
		LinkUrl:  fmt.Sprintf("%s/app/reset-password/%s", a.cfg.App.Url, token),
	}

	err = a.qs.SendMessage(emailData.ToString(), consttype.SEND_EMAIL)
	if err != nil {
		return err
	}

	return nil
}

func (a *AuthService) SendVerificationEmail(id uint, token int) error {
	user, err := a.userRepo.FindById(id)
	if err != nil {
		return err
	}

	if time.Now().UTC().Before(user.ConfirmationSentAt.Add(time.Minute * 5)) {
		return errors.New("you already requested a verification message in less than 5 minutes")
	}

	userUpdate := model.User{
		ConfirmationToken:  token,
		ConfirmationSentAt: time.Now().UTC(),
	}

	_, err = a.userRepo.Update(userUpdate, user.ID)
	if err != nil {
		return err
	}

	emailData := request.SendEmailRequest{
		Template: "verify_email.html",
		Subject:  "Verification Code",
		Name:     user.FullName,
		Email:    user.Email,
		Token:    token,
		LinkUrl:  "",
	}

	err = a.qs.SendMessage(emailData.ToString(), consttype.SEND_EMAIL)
	if err != nil {
		return err
	}

	return nil
}

func (a *AuthService) VerifyToken(req request.VerifyTokenRequest) error {
	user, err := a.userRepo.FindByEmail(req.Email)
	if err != nil {
		return err
	}

	if !time.Now().UTC().Before(user.ConfirmationSentAt.Add(time.Minute * 5)) {
		return errors.New("this token is expired")
	}

	if !user.ConfirmedAt.Equal(time.Time{}) {
		return errors.New("this user is already confirmed")
	}

	if req.Token != user.ConfirmationToken {
		return errors.New("this token is not the same")
	}

	userUpdate := model.User{
		ConfirmedAt: time.Now().UTC(),
	}

	_, err = a.userRepo.Update(userUpdate, user.ID)
	if err != nil {
		return err
	}

	return nil
}

func (a *AuthService) RefreshAuthToken(refreshToken string) (*response.UserResponse, *utils.TokenHeader, error) {
	parsedToken, err := utils.ParseToken(refreshToken, a.cfg.App.Secret)
	if err != nil {
		return nil, nil, err
	}

	user, err := a.userRepo.FindById(parsedToken.User.ID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil, errors.New("user not found")
		}

		return nil, nil, err
	}

	if time.Now().Unix() >= parsedToken.Expire {
		return nil, nil, errors.New("refresh token expired")
	}

	tokenHeader, err := a.generateAuthTokens(user)
	if err != nil {
		return nil, nil, err
	}

	return user, tokenHeader, err
}

func verifyPassword(u *response.UserResponse, password string) error {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))

	if err != nil {
		switch err {
		case bcrypt.ErrMismatchedHashAndPassword:
			return errors.New("password is incorrect")
		default:
			return err
		}
	}

	return err
}

func (a *AuthService) generateAuthTokens(user *response.UserResponse) (*utils.TokenHeader, error) {
	refreshToken, err := utils.GenerateToken(user, a.cfg.App.RefreshTokenLifespan, a.cfg.App.TokenLifespanDuration, a.cfg.App.Secret)
	if err != nil {
		return nil, err
	}

	token, err := utils.GenerateToken(user, a.cfg.App.TokenLifespan, a.cfg.App.TokenLifespanDuration, a.cfg.App.Secret)
	if err != nil {
		return nil, err
	}

	tokenHeader := utils.TokenHeader{
		AuthToken:           token.Token,
		AuthTokenExpires:    token.Expires,
		RefreshToken:        refreshToken.Token,
		RefreshTokenExpires: refreshToken.Expires,
	}

	return &tokenHeader, err
}
