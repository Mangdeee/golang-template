package repository

import (
	"github.com/felixlambertv/go-cleanplate/internal/controller/response"
	"github.com/felixlambertv/go-cleanplate/internal/model"
	"gorm.io/gorm"
)

// IUser

type (
	IUserRepo interface {
		WithTrx(trxHandle *gorm.DB) IUserRepo
		FindAll(p model.Pagination) (*model.Pagination, error)
		Store(user *model.User) (*model.User, error)
		Update(user model.User, userID uint) (*model.User, error)
		FindById(id uint) (*response.UserResponse, error)
		FindByEmail(email string) (*response.UserResponse, error)
		DeleteUser(user model.User) error
	}
)
