package validators

import (
	"errors"
	"strings"

	"tesina/backend/internal/models"
)

var (
	ErrPostalCodeRequired = errors.New("postal_code is required")
	ErrCityNameRequired   = errors.New("name is required")
	ErrAndenRequired      = errors.New("anden is required")
	ErrEmailRequired      = errors.New("email is required")
)

func ValidateCreateCityRequest(req models.CreateCityRequest) error {
	if req.PostalCode == "" {
		return ErrPostalCodeRequired
	}
	if req.Name == "" {
		return ErrCityNameRequired
	}
	return nil
}

func ValidateCreatePlatformRequest(req models.CreatePlatformRequest) error {
	if req.Anden == "" {
		return ErrAndenRequired
	}
	return nil
}

// ValidateAdminEmail trims the email and checks it is not empty.
// Returns the trimmed email on success.
func ValidateAdminEmail(email string) (string, error) {
	email = strings.TrimSpace(email)
	if email == "" {
		return "", ErrEmailRequired
	}
	return email, nil
}

func ValidatePromoteAdminRequest(req models.PromoteAdminRequest) error {
	if req.Email == "" {
		return ErrEmailRequired
	}
	return nil
}

func ValidateDemoteAdminRequest(req models.DemoteAdminRequest) error {
	if req.Email == "" {
		return ErrEmailRequired
	}
	return nil
}
