package response

import (
	"time"

	"gorm.io/gorm"
)

type (
	UserResponse struct {
		ID                     uint           `json:"id"`
		FullName               string         `json:"fullName" example:"user name"`
		Email                  string         `json:"email" example:"email@email.com"`
		Password               string         `json:"-" example:"password123"`
		UserLevel              uint           `json:"userLevel" example:"1"`
		Country                string         `json:"country" example:"country"`
		CountryCode            uint           `json:"countryCode" example:"62"`
		ScenarioCount          int            `json:"scenarioCount"`
		ResetPasswordToken     string         `json:"-"`
		ResetPasswordSentAt    time.Time      `json:"-"`
		ConfirmationToken      int            `json:"-"`
		ConfirmedAt            time.Time      `json:"confirmedAt"`
		ConfirmationSentAt     time.Time      `json:"-"`
		RefreshToken           string         `json:"-"`
		RefreshTokenExpiration string         `json:"-"`
		CreatedAt              time.Time      `json:"createdAt,omitempty" example:"2023-01-01T15:01:00+00:00"`
		UpdatedAt              time.Time      `json:"updatedAt,omitempty" example:"2023-02-11T15:01:00+00:00"`
		DeletedAt              gorm.DeletedAt `json:"-"`
	}
)
