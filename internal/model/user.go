package model

import (
	"time"

	"gorm.io/gorm"
)

type (
	User struct {
		ID                     uint           `gorm:"primary_key" json:"id"`
		FullName               string         `json:"fullName" gorm:"not null" example:"user name"`
		Email                  string         `json:"email" gorm:"not null;unique" example:"email@email.com"`
		Password               string         `json:"-" gorm:"not null" example:"password123"`
		UserLevel              uint           `json:"userLevel" gorm:"not null" example:"1"`
		Country                string         `json:"country" example:"country"`
		CountryCode            uint           `json:"countryCode" example:"62"`
		RefreshToken           string         `json:"-"`
		RefreshTokenExpiration string         `json:"-"`
		ResetPasswordToken     string         `json:"-"`
		ResetPasswordSentAt    time.Time      `json:"-"`
		ConfirmationToken      int            `json:"-"`
		ConfirmedAt            time.Time      `json:"-"`
		ConfirmationSentAt     time.Time      `json:"-"`
		CreatedAt              time.Time      `json:"createdAt,omitempty" example:"2023-01-01T15:01:00+00:00"`
		UpdatedAt              time.Time      `json:"updatedAt,omitempty" example:"2023-02-11T15:01:00+00:00"`
		DeletedAt              gorm.DeletedAt `json:"-"`
	}
)
