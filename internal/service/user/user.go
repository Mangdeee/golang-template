package user

import (
	"encoding/json"
	"fmt"

	"github.com/felixlambertv/go-cleanplate/internal/controller/request"
	"github.com/felixlambertv/go-cleanplate/internal/controller/response"
	"github.com/felixlambertv/go-cleanplate/internal/model"
	"github.com/felixlambertv/go-cleanplate/internal/repository"
	"github.com/felixlambertv/go-cleanplate/internal/service"
	"github.com/felixlambertv/go-cleanplate/pkg/utils"
	"gorm.io/gorm"
)

type UserService struct {
	userRepo repository.IUserRepo
}

func NewUserService(userRepo repository.IUserRepo) *UserService {
	return &UserService{userRepo: userRepo}
}

func (u *UserService) CreateUser(req request.CreateUserRequest) (*response.UserResponse, error) {
	var userResponse *response.UserResponse

	hashedPassword, err := utils.EncryptPassword(req.Password)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		FullName: req.Name,
		Email:    req.Email,
		Password: string(hashedPassword),
	}
	user, err = u.userRepo.Store(user)
	if err != nil {
		return nil, err
	}

	marshaledUser, _ := json.Marshal(user)
	err = json.Unmarshal(marshaledUser, &userResponse)
	if err != nil {
		fmt.Println("err", err)
	}

	return userResponse, err
}

func (u *UserService) UpdateUserCountry(req request.UpdateUserCountryRequest, userID uint) (*response.UserResponse, error) {
	var userResponse *response.UserResponse

	updateUserReq := model.User{
		Country: req.Country,
	}

	_, err := u.userRepo.Update(updateUserReq, userID)
	if err != nil {
		return nil, err
	}

	user, err := u.userRepo.FindById(userID)
	if err != nil {
		return nil, err
	}

	marshaledUser, _ := json.Marshal(user)
	err = json.Unmarshal(marshaledUser, &userResponse)
	if err != nil {
		return nil, err
	}

	return userResponse, err
}

func (u *UserService) GetUser(id uint) (*response.UserResponse, error) {
	user, err := u.userRepo.FindById(id)
	if err != nil {
		return nil, err
	}

	return user, err
}

func (u *UserService) GetUsers(paginationReq model.Pagination) (*model.Pagination, error) {
	users, err := u.userRepo.FindAll(paginationReq)
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (u *UserService) DeleteUser(id uint) error {
	userModel := model.User{
		ID: id,
	}
	err := u.userRepo.DeleteUser(userModel)
	if err != nil {
		return err
	}

	return err
}

func (u *UserService) WithTrx(trxHandle *gorm.DB) service.IUserService {
	u.userRepo = u.userRepo.WithTrx(trxHandle)
	return u
}
