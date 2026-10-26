package validators

import (
	"errors"
	"strings"

	"tesina/backend/internal/models"
)

var ErrTokenRequired = errors.New("token is required")

func ValidateForgotPasswordRequest(req models.ForgotPasswordRequest) error {
	email := strings.TrimSpace(req.Email)
	if email == "" {
		return ErrEmailRequired
	}
	return validateEmailFormat(email)
}

func ValidateRecoveryTokenRequest(token string) error {
	if strings.TrimSpace(token) == "" {
		return ErrTokenRequired
	}
	return nil
}

func ValidateResetPasswordWithTokenRequest(req models.ResetPasswordRequest) error {
	if strings.TrimSpace(req.Password) == "" {
		return ErrPasswordRequired
	}
	return nil
}
