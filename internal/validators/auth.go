package validators

import (
	"errors"

	"tesina/backend/internal/models"
)

var (
	ErrPasswordRequired     = errors.New("password is required")
	ErrFirstNameRequired    = errors.New("first_name is required")
	ErrLastNameRequired     = errors.New("last_name is required")
	ErrDNIRequired          = errors.New("dni is required")
	ErrRefreshTokenRequired = errors.New("refresh_token is required")
)

func ValidateCreateUserRequest(req models.CreateUserRequest) error {
	if req.Email == "" {
		return ErrEmailRequired
	}
	if req.Password == "" {
		return ErrPasswordRequired
	}
	if req.FirstName == "" {
		return ErrFirstNameRequired
	}
	if req.LastName == "" {
		return ErrLastNameRequired
	}
	if req.DNI == "" {
		return ErrDNIRequired
	}
	return nil
}

func ValidateLoginRequest(req models.LoginRequest) error {
	if req.Email == "" {
		return ErrEmailRequired
	}
	if req.Password == "" {
		return ErrPasswordRequired
	}
	return nil
}

func ValidateRefreshTokenRequest(req models.RefreshTokenRequest) error {
	if req.RefreshToken == "" {
		return ErrRefreshTokenRequired
	}
	return nil
}

func ValidateLogoutRequest(req models.LogoutRequest) error {
	if req.RefreshToken == "" {
		return ErrRefreshTokenRequired
	}
	return nil
}
