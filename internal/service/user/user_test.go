package user

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/felixlambertv/go-cleanplate/internal/controller/request"
	"github.com/felixlambertv/go-cleanplate/internal/controller/response"
	"github.com/felixlambertv/go-cleanplate/internal/model"
	"github.com/felixlambertv/go-cleanplate/mocks"
	"github.com/felixlambertv/go-cleanplate/pkg/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var userRepoMock = new(mocks.IUserRepo)
var userService = NewUserService(userRepoMock)

var paginationRequest = &model.Pagination{
	Limit: 10,
	Page:  1,
	Sort:  "id asc",
	Filter: model.Filter{
		CreatedFrom: time.Date(int(2023), time.March, int(27), int(0), int(0), int(0), int(0), &time.Location{}),
		CreatedTo:   time.Date(int(2023), time.March, int(31), int(0), int(0), int(0), int(0), &time.Location{}),
	},
}

var paginationDummy = &model.Pagination{
	Limit:      10,
	Page:       1,
	Sort:       "id asc",
	TotalDatas: 0,
	TotalPages: 1,
	Data:       []model.User{},
}

var createUserRequest = request.CreateUserRequest{
	Name:     "test",
	Email:    "test@example.com",
	Password: "password",
}

var userDummy = &model.User{
	ID:          uint(1),
	FullName:    "test",
	Email:       "test@example.com",
	Password:    "password",
	UserLevel:   0,
	Country:     "indonesia",
	CountryCode: 0,
}

var userResponseDummy = &response.UserResponse{
	ID:            1,
	FullName:      "test",
	Email:         "test@example.com",
	UserLevel:     0,
	Country:       "indonesia",
	CountryCode:   0,
	ScenarioCount: 0,
}

var updateUserCountryRequest = request.UpdateUserCountryRequest{
	Country: "USA",
}

var updatedUserResponseDummy = &response.UserResponse{
	ID:            1,
	FullName:      "test",
	Email:         "test@example.com",
	UserLevel:     0,
	Country:       "USA",
	CountryCode:   0,
	ScenarioCount: 0,
}

var updatedUserDummy = &model.User{
	ID:          1,
	FullName:    "test",
	Email:       "test@example.com",
	Password:    "password",
	UserLevel:   0,
	Country:     "USA",
	CountryCode: 0,
}

func TestMain(m *testing.M) {
	fmt.Print("before")

	encryptedPassword, _ := utils.EncryptPassword("password")
	createUserRequest.Password = string(encryptedPassword)
	userDummy.Password = string(encryptedPassword)
	updatedUserDummy.Password = string(encryptedPassword)

	m.Run()
	fmt.Println("after")
}

func TestUser_GetUsers(t *testing.T) {
	userRepoMock.On("FindAll", *paginationRequest).Return(paginationDummy, nil)
	users, err := userService.GetUsers(*paginationRequest)
	if err != nil {
		fmt.Println(err)
	}

	assert.Equal(t, paginationDummy, users)
}

func TestUser_GetUser(t *testing.T) {
	userRepoMock.On("FindById", userDummy.ID).Return(userResponseDummy, nil)
	user, err := userService.GetUser(userDummy.ID)
	if err != nil {
		fmt.Println(err)
	}

	assert.Equal(t, user, userResponseDummy)
	assert.Nil(t, err)
}

func TestUser_CreateUserSuccessful(t *testing.T) {
	userRepoMock.ExpectedCalls = nil
	userRepoMock.On("Store", mock.Anything).Return(userDummy, nil)

	user, err := userService.CreateUser(createUserRequest)
	if err != nil {
		fmt.Println(err)
	}

	assert.Equal(t, userResponseDummy, user)
	assert.Nil(t, err)
}

func TestUser_CreateUserShouldReturnError(t *testing.T) {
	userRepoMock.ExpectedCalls = nil
	userRepoMock.On("Store", mock.Anything).Return(nil, errors.New("something went wrong"))

	user, err := userService.CreateUser(createUserRequest)
	if err != nil {
		fmt.Println(err)
	}

	assert.Nil(t, user)
	assert.Equal(t, errors.New("something went wrong"), err)
}

func TestScenario_UpdateUserCountrySuccessful(t *testing.T) {
	userRepoMock.ExpectedCalls = nil

	userRepoMock.On("Update", model.User{Country: "USA"}, userResponseDummy.ID).Return(updatedUserDummy, nil).Once()
	userRepoMock.On("FindById", userDummy.ID).Return(updatedUserResponseDummy, nil).Once()

	user, err := userService.UpdateUserCountry(updateUserCountryRequest, userResponseDummy.ID)
	if err != nil {
		fmt.Println(err)
	}

	assert.Equal(t, updatedUserResponseDummy, user)
	assert.Nil(t, err)
}

func TestUser_DeleteUserSuccessful(t *testing.T) {
	userRepoMock.On("DeleteUser", model.User{ID: userDummy.ID}).Return(nil)

	err := userService.DeleteUser(userDummy.ID)
	if err != nil {
		fmt.Println(err)
	}

	assert.Nil(t, err)
}

func TestUser_DeleteUserError(t *testing.T) {
	userRepoMock.On("DeleteUser", model.User{ID: uint(2)}).Return(errors.New("something went wrong"))

	err := userService.DeleteUser(uint(2))
	if err != nil {
		fmt.Println(err)
	}

	assert.Equal(t, errors.New("something went wrong"), err)
}
