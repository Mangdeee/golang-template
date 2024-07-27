package request

import (
	"encoding/json"
	"fmt"
)

type (
	SendEmailRequest struct {
		Template string `validate:"required,oneof=reset_password.html verify_email.html"`
		Subject  string `validate:"required"`
		Name     string
		Email    string `validate:"required,email"`
		Token    int
		LinkUrl  string
	}
)

func (s SendEmailRequest) ToString() string {
	b, err := json.Marshal(s)
	if err != nil {
		fmt.Printf("Error: %s", err)
		return ""
	}
	return string(b)
}
