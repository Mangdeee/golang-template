package user

import (
	"fmt"

	"github.com/felixlambertv/go-cleanplate/internal/controller/response"
	"github.com/felixlambertv/go-cleanplate/internal/model"
	"github.com/felixlambertv/go-cleanplate/internal/repository"
	"github.com/felixlambertv/go-cleanplate/internal/repository/pagination"
	"github.com/felixlambertv/go-cleanplate/pkg/consttype"
	"github.com/felixlambertv/go-cleanplate/pkg/logger"
	"gorm.io/gorm"
)

type UserRepo struct {
	l  logger.Interface
	db *gorm.DB
}

func NewUserRepo(db *gorm.DB, l logger.Interface) *UserRepo {
	return &UserRepo{db: db, l: l}
}

func (u *UserRepo) WithTrx(trxHandle *gorm.DB) repository.IUserRepo {
	if trxHandle == nil {
		u.l.Error("transaction db not found")
		return u
	}
	u.db = trxHandle
	return u
}

func (u *UserRepo) Store(user *model.User) (*model.User, error) {
	err := u.db.Create(user).Error
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (u *UserRepo) Update(user model.User, userID uint) (*model.User, error) {
	err := u.db.Model(&user).Where("id = ?", userID).Updates(user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (u *UserRepo) FindAll(p model.Pagination) (*model.Pagination, error) {
	var users []model.User
	var usersResponse []response.UserResponse

	result := u.db.Model(&users).Select("users.id as id, full_name, email, password, user_level, country, country_code, reset_password_token, reset_password_sent_at, confirmation_token, confirmed_at, confirmation_sent_at, users.created_at as created_at,refresh_token, refresh_token_expiration, users.updated_at as updated_at")

	if p.Search != "" {
		result = result.Where("full_name LIKE ?", fmt.Sprintf("%%%s%%", p.Search)).Or("email LIKE ?", fmt.Sprintf("%%%s%%", p.Search))
	}

	if !p.Filter.CreatedFrom.IsZero() && !p.Filter.CreatedTo.IsZero() {
		result = result.Where("date(users.created_at) between ? and ?", p.Filter.CreatedFrom.Format(consttype.DATEFORMAT), p.Filter.CreatedTo.Format(consttype.DATEFORMAT))
	}

	result = result.Group("users.id").Scopes(pagination.Paginate(&users, &p, result)).Find(&usersResponse)

	if result.Error != nil {
		return &p, result.Error
	}

	p.Data = usersResponse
	return &p, nil
}

func (u *UserRepo) FindById(id uint) (*response.UserResponse, error) {
	var user *response.UserResponse
	err := u.db.Model(&model.User{}).Select("users.id as id, full_name, email, password, user_level, country, country_code, reset_password_token, reset_password_sent_at, confirmation_token, confirmed_at, confirmation_sent_at, refresh_token, refresh_token_expiration, users.created_at as created_at, users.updated_at as updated_at").Group("users.id").First(&user, id).Error
	if err != nil {
		return nil, err
	}

	return user, err
}

func (u *UserRepo) FindByEmail(email string) (*response.UserResponse, error) {
	var user *response.UserResponse
	err := u.db.Model(&model.User{}).Select("users.id as id, full_name, email, password, user_level, country, country_code, reset_password_token, reset_password_sent_at, confirmation_token, confirmed_at, confirmation_sent_at, refresh_token, refresh_token_expiration, users.created_at as created_at, users.updated_at as updated_at").Where("email = ?", email).Group("users.id").Take(&user).Error
	if err != nil {
		return nil, err
	}

	return user, err
}

func (u *UserRepo) DeleteUser(user model.User) error {
	err := u.db.Unscoped().Delete(&user).Error
	if err != nil {
		return err
	}

	return nil
}
