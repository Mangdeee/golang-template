package request

type (
	LoginRequest struct {
		Email    string `json:"email" binding:"required,email" example:"email@email.com"`
		Password string `json:"password" binding:"required" example:"password123"`
	}

	RegisterRequest struct {
		FullName string `json:"fullName" binding:"required" example:"Full Name"`
		Email    string `json:"email" binding:"required,email" example:"email@email.com"`
		Password string `json:"password" binding:"required" example:"password123"`
		Country  string `json:"country" binding:"required" example:"indonesia"`
	}

	VerifyTokenRequest struct {
		Token int    `json:"token" binding:"required,number" example:""`
		Email string `json:"email" binding:"required,email"`
	}

	ForgotPasswordRequest struct {
		Email string `json:"email" binding:"required,email"`
	}

	ResetPasswordRequest struct {
		Email           string `json:"email" binding:"required,email"`
		Password        string `json:"password" binding:"required"`
		ConfirmPassword string `json:"confirmPassword" binding:"required,eqfield=Password"`
		Token           string `json:"token" binding:"required"`
	}

	ResetPasswordRedirectRequest struct {
		ResetToken string `uri:"token" binding:"required"`
	}
)
