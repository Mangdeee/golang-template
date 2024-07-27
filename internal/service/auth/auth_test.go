package auth

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/felixlambertv/go-cleanplate/config"
	"github.com/felixlambertv/go-cleanplate/internal/controller/request"
	"github.com/felixlambertv/go-cleanplate/internal/controller/response"
	"github.com/felixlambertv/go-cleanplate/internal/model"
	"github.com/felixlambertv/go-cleanplate/mocks"
	"github.com/felixlambertv/go-cleanplate/pkg/consttype"
	"github.com/felixlambertv/go-cleanplate/pkg/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

var cfg = &config.Config{
	App: config.App{
		Secret:                "randomblabla",
		TokenLifespan:         1,
		TokenLifespanDuration: "minute",
		DeeplinkUrl:           "test://",
	},
	Mail: config.Mail{
		Host:     "",
		Port:     0,
		User:     "",
		Password: "",
		From:     "no-reply@test.com",
	},
}
var userRepoMock = new(mocks.IUserRepo)
var mailServiceMock = new(mocks.IMailService)
var queueServiceMock = new(mocks.IQueueService)

var authService = NewAuthService(userRepoMock, cfg, mailServiceMock, queueServiceMock)

var VerifyTokenRequest = request.VerifyTokenRequest{
	Email: "user@test.com",
	Token: 8128,
}

var ResetPasswordRequest = request.ResetPasswordRequest{
	Email:           "user@test.com",
	Password:        "newpassword",
	ConfirmPassword: "newpassword",
	Token:           "2l5hlPdxEdSi9bT5",
}

var LoginRequest = request.LoginRequest{
	Email:    "user@test.com",
	Password: "password",
}

var ErrorLoginRequest = request.LoginRequest{
	Email:    "user@example.com",
	Password: "Password",
}

var userResponseDummy = &response.UserResponse{
	ID:                     0,
	Email:                  "user@test.com",
	Password:               "password",
	FullName:               "Test User",
	UserLevel:              consttype.USER,
	Country:                "indonesia",
	CountryCode:            62,
	ScenarioCount:          0,
	ConfirmationToken:      8128,
	ConfirmationSentAt:     time.Now().Add(time.Minute * time.Duration(-5)),
	ResetPasswordToken:     "2l5hlPdxEdSi9bT5",
	ResetPasswordSentAt:    time.Now().Add(time.Minute * time.Duration(-5)),
	RefreshToken:           "testtest123",
	RefreshTokenExpiration: time.Now().Add(time.Minute * time.Duration(60)).Format(time.RFC3339),
	ConfirmedAt:            time.Time{},
	CreatedAt:              time.Time{},
	UpdatedAt:              time.Time{},
	DeletedAt:              gorm.DeletedAt{},
}

var userDummy = model.User{
	ID:                  0,
	Email:               "user@test.com",
	Password:            "password",
	FullName:            "Test User",
	UserLevel:           consttype.USER,
	Country:             "indonesia",
	CountryCode:         62,
	ConfirmationToken:   8128,
	ConfirmationSentAt:  time.Now().Add(time.Minute * time.Duration(-5)),
	ResetPasswordToken:  "2l5hlPdxEdSi9bT5",
	ResetPasswordSentAt: time.Now().Add(time.Minute * time.Duration(-5)),
	ConfirmedAt:         time.Time{},
	CreatedAt:           time.Time{},
	UpdatedAt:           time.Time{},
	DeletedAt:           gorm.DeletedAt{},
}

var notFoundEmailRegisterRequest = request.RegisterRequest{
	FullName: "Test User",
	Email:    "user@test.com",
	Password: "password",
	Country:  "indonesia",
}

var errorStoreRegisterRequest = request.RegisterRequest{
	FullName: "Test User",
	Email:    "user@test.com",
	Password: "password",
	Country:  "indonesia",
}

var RegisterRequest = request.RegisterRequest{
	FullName: "Test User",
	Email:    "user@TOAST.com",
	Password: "password",
	Country:  "indonesia",
}

var userRegisterDummy = &model.User{
	ID:          0,
	Email:       "user@toast.com",
	Password:    "password",
	FullName:    "Test User",
	UserLevel:   consttype.USER,
	Country:     "indonesia",
	CountryCode: 62,
	CreatedAt:   time.Time{},
	UpdatedAt:   time.Time{},
	DeletedAt:   gorm.DeletedAt{},
}

var userRegisterResult = &response.UserResponse{
	ID:            0,
	Email:         "user@toast.com",
	FullName:      "Test User",
	UserLevel:     consttype.USER,
	Country:       "indonesia",
	CountryCode:   62,
	ScenarioCount: 0,
	CreatedAt:     time.Time{},
	UpdatedAt:     time.Time{},
	DeletedAt:     gorm.DeletedAt{},
}

var emailData = request.SendEmailRequest{
	Template: "verify_email.html",
	Subject:  "Verification Code",
	Name:     userDummy.FullName,
	Email:    userDummy.Email,
	Token:    8128,
	LinkUrl:  "",
}

var resetPasswordEmailData = request.SendEmailRequest{
	Template: "reset_password.html",
	Subject:  "Reset Password",
	Name:     userDummy.FullName,
	Email:    userDummy.Email,
	Token:    0,
	LinkUrl:  fmt.Sprintf("%sreset-password/%s", cfg.DeeplinkUrl, "2l5hlPdxEdSi9bT5"),
}

var updateUserRequest = model.User{
	ConfirmationToken:   8128,
	ConfirmationSentAt:  userDummy.ConfirmationSentAt,
	ConfirmedAt:         time.Time{},
	ResetPasswordSentAt: userDummy.ResetPasswordSentAt,
	ResetPasswordToken:  "",
}

func BeforeEachVerificationTest(confirmationSentAt time.Time, confirmedAt time.Time) {
	userResponseDummy.ConfirmationSentAt = confirmationSentAt
	userResponseDummy.ConfirmedAt = confirmedAt
	userResponseDummy.ResetPasswordSentAt = confirmationSentAt
	userResponseDummy.ResetPasswordToken = "2l5hlPdxEdSi9bT5"

	mailServiceMock.ExpectedCalls = nil
	userRepoMock.ExpectedCalls = nil
}

func TestMain(m *testing.M) {
	fmt.Print("before")

	generatedRandomToken := utils.GenerateRandomToken()
	uDummyHashed, err := utils.EncryptPassword(userDummy.Password)
	if err != nil {
		log.Fatal("error on encrypting password", err)
	}

	refreshToken, err := utils.GenerateToken(userResponseDummy, cfg.App.TokenLifespan+1, cfg.App.TokenLifespanDuration, cfg.App.Secret)
	if err != nil {
		fmt.Println("error on generating token", err)
	}

	userResponseDummy.RefreshToken = refreshToken.Token
	userResponseDummy.RefreshTokenExpiration = refreshToken.Expires.String()

	userDummy.Password = uDummyHashed
	userResponseDummy.Password = uDummyHashed

	updateUserRequest.ConfirmationToken = generatedRandomToken
	VerifyTokenRequest.Token = generatedRandomToken
	userResponseDummy.ConfirmationToken = generatedRandomToken
	userDummy.ConfirmationToken = generatedRandomToken
	emailData.Token = generatedRandomToken

	m.Run()
	fmt.Println("after")
}

func TestAuth_LoginSuccess(t *testing.T) {
	userRepoMock.On("Update", mock.Anything, userDummy.ID).Return(&userDummy, nil).Once()
	userRepoMock.On("FindByEmail", userDummy.Email).Return(userResponseDummy, nil).Once()

	user, token, err := authService.Login(LoginRequest)
	if err != nil {
		fmt.Println(err)
	}
	assert.Equal(t, user, userResponseDummy)
	assert.Equal(t, assert.NotNil(t, token), true)
	assert.Equal(t, err, nil)
}

func TestAuth_LoginShouldReturnEmailNotFound(t *testing.T) {
	userRepoMock.On("Update", mock.Anything, userDummy.ID).Return(&userDummy, nil).Once()
	userRepoMock.On("FindByEmail", ErrorLoginRequest.Email).Return(nil, gorm.ErrRecordNotFound).Once()

	user, token, err := authService.Login(ErrorLoginRequest)
	if err != nil {
		fmt.Println(err)
	}
	assert.Nil(t, user)
	assert.Nil(t, token)
	assert.Equal(t, "user not found", err.Error())
}

func TestAuth_LoginShouldReturnInvalidPassword(t *testing.T) {
	userRepoMock.On("Update", mock.Anything, userDummy.ID).Return(&userDummy, nil).Once()
	userRepoMock.On("FindByEmail", userDummy.Email).Return(userResponseDummy, nil).Once()

	user, token, err := authService.Login(request.LoginRequest{
		Email:    userDummy.Email,
		Password: ErrorLoginRequest.Password,
	})
	if err != nil {
		fmt.Println(err)
	}

	assert.Nil(t, user)
	assert.Nil(t, token)
	assert.Equal(t, "password is incorrect", err.Error())
}

func TestAuth_RegisterSuccess(t *testing.T) {
	userRepoMock.On("FindByEmail", strings.ToLower(RegisterRequest.Email)).Return(nil, gorm.ErrRecordNotFound).Once()
	userRepoMock.On("Store", mock.Anything).Return(userRegisterDummy, nil).Once()

	user, token, err := authService.Register(RegisterRequest)
	if err != nil {
		fmt.Println(err)
	}
	assert.Equal(t, userRegisterResult, user)
	assert.Equal(t, assert.NotNil(t, token), true)
	assert.Equal(t, err, nil)
}

func TestAuth_RegisterShouldReturnEmailFound(t *testing.T) {
	userRepoMock.On("FindByEmail", notFoundEmailRegisterRequest.Email).Return(userResponseDummy, nil).Once()
	userRepoMock.On("Store", mock.Anything).Return(nil, errors.New("something went wrong")).Once()

	user, token, err := authService.Register(notFoundEmailRegisterRequest)
	if err != nil {
		fmt.Println(err)
	}

	assert.Nil(t, user)
	assert.Nil(t, token)
	assert.Equal(t, "user already exists", err.Error())
}

func TestAuth_RegisterShouldReturnStoreError(t *testing.T) {
	userRepoMock.On("FindByEmail", errorStoreRegisterRequest.Email).Return(nil, nil).Once()
	userRepoMock.On("Store", mock.Anything).Return(nil, errors.New("something went wrong")).Once()

	user, token, err := authService.Register(errorStoreRegisterRequest)

	assert.Nil(t, user)
	assert.Nil(t, token)
	assert.Equal(t, "something went wrong", err.Error())
}

func TestAuth_SendVerificationEmailSuccessful(t *testing.T) {
	userRepoMock.On("FindById", userDummy.ID).Return(userResponseDummy, nil).Once()
	userRepoMock.On("Update", mock.Anything, userDummy.ID).Return(&userDummy, nil).Once()
	queueServiceMock.On("SendMessage", mock.Anything, mock.Anything).Return(nil).Once()

	err := authService.SendVerificationEmail(userDummy.ID, emailData.Token)

	assert.Nil(t, err)
}

func TestAuth_SendVerificationEmailShouldReturnUserNotFound(t *testing.T) {
	userRepoMock.On("FindById", userDummy.ID).Return(nil, gorm.ErrRecordNotFound).Once()
	userRepoMock.On("Update", mock.Anything, userDummy.ID).Return(&userDummy, nil).Once()

	err := authService.SendVerificationEmail(userDummy.ID, emailData.Token)

	assert.Equal(t, err, gorm.ErrRecordNotFound)
}

func TestAuth_SendVerificationEmailShouldReturnSendEmailError(t *testing.T) {
	BeforeEachVerificationTest(time.Now().UTC().Add(time.Minute*time.Duration(-5)), time.Time{})

	userRepoMock.On("FindById", userDummy.ID).Return(userResponseDummy, nil).Once()
	queueServiceMock.On("SendMessage", mock.Anything, mock.Anything).Return(errors.New("send to queue error")).Once()
	userRepoMock.On("Update", mock.Anything, userDummy.ID).Return(&userDummy, nil).Once()

	err := authService.SendVerificationEmail(userDummy.ID, emailData.Token)

	assert.Equal(t, err, errors.New("send to queue error"))
}

func TestAuth_SendVerificationEmailShouldReturnUpdateError(t *testing.T) {
	BeforeEachVerificationTest(time.Now().UTC().Add(time.Minute*time.Duration(-5)), time.Time{})

	userRepoMock.On("FindById", userDummy.ID).Return(userResponseDummy, nil).Once()
	mailServiceMock.On("SendEmail", emailData).Return(nil).Once()
	userRepoMock.On("Update", mock.Anything, userDummy.ID).Return(nil, errors.New("update went wrong")).Once()

	err := authService.SendVerificationEmail(userDummy.ID, emailData.Token)

	assert.Equal(t, err, errors.New("update went wrong"))
}

func TestAuth_SendVerificationEmailShouldReturnAlreadyRequestedError(t *testing.T) {
	BeforeEachVerificationTest(time.Now().UTC().Add(time.Minute*2), time.Time{})

	userRepoMock.On("FindById", userDummy.ID).Return(userResponseDummy, nil).Once()
	mailServiceMock.On("SendEmail", emailData).Return(errors.New("you already requested a verification message in less than 5 minutes")).Once()
	userRepoMock.On("Update", mock.Anything, userDummy.ID).Return(nil, errors.New("something update went wrong")).Once()

	err := authService.SendVerificationEmail(userDummy.ID, emailData.Token)

	assert.Equal(t, err, errors.New("you already requested a verification message in less than 5 minutes"))
}

func TestAuth_VerifyTokenSuccessful(t *testing.T) {
	BeforeEachVerificationTest(time.Now().UTC().Add(time.Minute*time.Duration(-4)), time.Time{})

	userRepoMock.On("FindByEmail", userDummy.Email).Return(userResponseDummy, nil).Once()
	userRepoMock.On("Update", mock.Anything, userDummy.ID).Return(&userDummy, nil).Once()

	err := authService.VerifyToken(VerifyTokenRequest)

	assert.Nil(t, err)
}

func TestAuth_VerifyTokenShouldReturnTokenExpired(t *testing.T) {
	BeforeEachVerificationTest(time.Now().UTC().Add(time.Minute*time.Duration(-5)), time.Time{})

	userRepoMock.On("FindByEmail", userDummy.Email).Return(userResponseDummy, nil).Once()
	userRepoMock.On("Update", mock.Anything, userDummy.ID).Return(&userDummy, nil).Once()

	err := authService.VerifyToken(VerifyTokenRequest)

	assert.Equal(t, err, errors.New("this token is expired"))
}

func TestAuth_VerifyTokenShouldReturnTokenAlreadyConfirmed(t *testing.T) {
	BeforeEachVerificationTest(time.Now().UTC().Add(time.Minute*time.Duration(-4)), time.Now().UTC().Add(time.Minute*time.Duration(2)))

	userRepoMock.On("FindByEmail", userDummy.Email).Return(userResponseDummy, nil).Once()
	userRepoMock.On("Update", mock.Anything, userDummy.ID).Return(&userDummy, nil).Once()

	err := authService.VerifyToken(VerifyTokenRequest)

	assert.Equal(t, err, errors.New("this user is already confirmed"))
}

func TestAuth_VerifyTokenShouldReturnUpdateError(t *testing.T) {
	BeforeEachVerificationTest(time.Now().UTC().Add(time.Minute*time.Duration(-4)), time.Time{})

	userRepoMock.On("FindByEmail", userDummy.Email).Return(userResponseDummy, nil).Once()
	userRepoMock.On("Update", mock.Anything, userDummy.ID).Return(nil, errors.New("update went wrong")).Once()

	err := authService.VerifyToken(VerifyTokenRequest)

	assert.Equal(t, err, errors.New("update went wrong"))
}

func TestAuth_VerifyTokenShouldReturnTokenNotSame(t *testing.T) {
	BeforeEachVerificationTest(time.Now().UTC().Add(time.Minute*time.Duration(-4)), time.Time{})

	userRepoMock.On("FindByEmail", userDummy.Email).Return(userResponseDummy, nil).Once()
	userRepoMock.On("Update", mock.Anything, userDummy.ID).Return(&userDummy, nil).Once()

	err := authService.VerifyToken(request.VerifyTokenRequest{
		Email: "user@test.com",
		Token: 1234,
	})

	assert.Equal(t, err, errors.New("this token is not the same"))
}

func TestAuth_ForgotPasswordSuccessful(t *testing.T) {
	BeforeEachVerificationTest(time.Now().UTC().Add(time.Minute*time.Duration(-10)), time.Time{})
	userRepoMock.On("FindByEmail", userDummy.Email).Return(userResponseDummy, nil).Once()
	userRepoMock.On("FindById", userDummy.ID).Return(userResponseDummy, nil).Once()
	userRepoMock.On("Update", mock.Anything, userDummy.ID).Return(&userDummy, nil).Once()
	queueServiceMock.On("SendMessage", mock.Anything, mock.Anything).Return(nil).Once()

	err := authService.ForgotPassword(request.ForgotPasswordRequest{
		Email: "user@test.com",
	})

	assert.Nil(t, err)
}

func TestAuth_ForgotPasswordShouldErrorUserNotFound(t *testing.T) {
	BeforeEachVerificationTest(time.Now().UTC().Add(time.Minute*time.Duration(-10)), time.Time{})
	userRepoMock.On("FindByEmail", userDummy.Email).Return(nil, gorm.ErrRecordNotFound).Once()
	userRepoMock.On("FindById", userDummy.ID).Return(nil, gorm.ErrRecordNotFound).Once()
	userRepoMock.On("Update", mock.Anything, userDummy.ID).Return(&userDummy, nil).Once()
	mailServiceMock.On("SendEmail", mock.Anything).Return(nil).Once()

	err := authService.ForgotPassword(request.ForgotPasswordRequest{
		Email: "user@test.com",
	})

	assert.Equal(t, "user not found", err.Error())
}

func TestAuth_ResetPasswordSuccessful(t *testing.T) {
	BeforeEachVerificationTest(time.Now().UTC().Add(time.Minute*time.Duration(-4)), time.Time{})
	userRepoMock.On("FindByEmail", userDummy.Email).Return(userResponseDummy, nil).Once()
	userRepoMock.On("Update", mock.Anything, userDummy.ID).Return(&userDummy, nil).Once()

	err := authService.ResetPassword(ResetPasswordRequest)

	assert.Nil(t, err)
}

func TestAuth_ResetPasswordShouldErrorUserNotFound(t *testing.T) {
	BeforeEachVerificationTest(time.Now().UTC().Add(time.Minute*time.Duration(-4)), time.Time{})
	userRepoMock.On("FindByEmail", userDummy.Email).Return(nil, gorm.ErrRecordNotFound).Once()
	userRepoMock.On("Update", mock.Anything, userDummy.ID).Return(&userDummy, nil).Once()

	err := authService.ResetPassword(ResetPasswordRequest)

	assert.Equal(t, "user not found", err.Error())
}

func TestAuth_ResetPasswordShouldReturnUpdateError(t *testing.T) {
	BeforeEachVerificationTest(time.Now().UTC().Add(time.Minute*time.Duration(-4)), time.Time{})
	userRepoMock.ExpectedCalls = nil
	userRepoMock.On("FindByEmail", userDummy.Email).Return(userResponseDummy, nil).Once()
	userRepoMock.On("Update", mock.Anything, userDummy.ID).Return(nil, errors.New("update went wrong")).Once()

	err := authService.ResetPassword(ResetPasswordRequest)

	assert.Equal(t, err, errors.New("update went wrong"))
}

func TestAuth_ResetPasswordShouldReturnTokenExpired(t *testing.T) {
	BeforeEachVerificationTest(time.Now().UTC().Add(time.Minute*time.Duration(-5)), time.Time{})

	userRepoMock.On("FindByEmail", userDummy.Email).Return(userResponseDummy, nil).Once()
	userRepoMock.On("Update", mock.Anything, userDummy.ID).Return(&userDummy, nil).Once()

	err := authService.ResetPassword(ResetPasswordRequest)

	assert.Equal(t, err, errors.New("this token is expired"))
}

func TestAuth_SendResetPasswordEmailSuccessful(t *testing.T) {
	userRepoMock.On("FindById", userDummy.ID).Return(userResponseDummy, nil).Once()
	userRepoMock.On("Update", mock.Anything, userDummy.ID).Return(&userDummy, nil).Once()
	queueServiceMock.On("SendMessage", mock.Anything, mock.Anything).Return(nil).Once()

	err := authService.SendResetPasswordEmail(userDummy.ID, userDummy.ResetPasswordToken)

	assert.Nil(t, err)
}

func TestAuth_SendResetPasswordEmailShouldReturnUserNotFound(t *testing.T) {
	userRepoMock.On("FindById", userDummy.ID).Return(nil, gorm.ErrRecordNotFound).Once()
	mailServiceMock.On("SendEmail", resetPasswordEmailData).Return(nil).Once()
	userRepoMock.On("Update", mock.Anything, userDummy.ID).Return(&userDummy, nil).Once()

	err := authService.SendResetPasswordEmail(userDummy.ID, userDummy.ResetPasswordToken)

	assert.Equal(t, err, gorm.ErrRecordNotFound)
}

func TestAuth_SendResetPasswordEmailShouldReturnSendEmailError(t *testing.T) {
	BeforeEachVerificationTest(time.Now().UTC().Add(time.Minute*time.Duration(-5)), time.Time{})

	userRepoMock.On("FindById", userDummy.ID).Return(userResponseDummy, nil).Once()
	userRepoMock.On("Update", mock.Anything, userDummy.ID).Return(&userDummy, nil).Once()
	queueServiceMock.On("SendMessage", mock.Anything, mock.Anything).Return(errors.New("sending email went wrong")).Once()

	err := authService.SendResetPasswordEmail(userDummy.ID, userDummy.ResetPasswordToken)

	assert.Equal(t, err, errors.New("sending email went wrong"))
}

func TestAuth_SendResetPasswordEmailShouldReturnUpdateError(t *testing.T) {
	BeforeEachVerificationTest(time.Now().UTC().Add(time.Minute*time.Duration(-5)), time.Time{})

	userRepoMock.On("FindById", userDummy.ID).Return(userResponseDummy, nil).Once()
	mailServiceMock.On("SendEmail", resetPasswordEmailData).Return(nil).Once()
	userRepoMock.On("Update", mock.Anything, userDummy.ID).Return(nil, errors.New("update went wrong")).Once()

	err := authService.SendResetPasswordEmail(userDummy.ID, userDummy.ResetPasswordToken)

	assert.Equal(t, err, errors.New("update went wrong"))
}

func TestAuth_SendResetPasswordEmailShouldReturnAlreadyRequestedError(t *testing.T) {
	BeforeEachVerificationTest(time.Now().UTC().Add(time.Minute*2), time.Time{})

	userRepoMock.On("FindById", userDummy.ID).Return(userResponseDummy, nil).Once()
	mailServiceMock.On("SendEmail", resetPasswordEmailData).Return(errors.New("you already requested a verification message in less than 5 minutes")).Once()
	userRepoMock.On("Update", mock.Anything, userDummy.ID).Return(nil, errors.New("something update went wrong")).Once()

	err := authService.SendResetPasswordEmail(userDummy.ID, userDummy.ResetPasswordToken)

	assert.Equal(t, err, errors.New("you already requested a reset password email in less than 5 minutes"))
}

func TestAuth_RefreshAuthTokenSuccessful(t *testing.T) {
	BeforeEachVerificationTest(time.Now().UTC().Add(time.Minute*2), time.Time{})

	userRepoMock.On("FindById", userDummy.ID).Return(userResponseDummy, nil).Once()
	userRepoMock.On("Update", mock.Anything, userDummy.ID).Return(&userDummy, nil).Once()

	user, token, err := authService.RefreshAuthToken(userResponseDummy.RefreshToken)

	assert.Equal(t, user, userResponseDummy)
	assert.Equal(t, assert.NotNil(t, token), true)
	assert.Equal(t, err, nil)
}
