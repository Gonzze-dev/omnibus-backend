package validators

import (
	"errors"
	"regexp"
	"strings"

	"tesina/backend/internal/models"
)

var (
	ErrPostalCodeRequired = errors.New("postal_code is required")
	ErrCityNameRequired   = errors.New("name is required")
	ErrBayRequired      = errors.New("bay is required")
	ErrBayInvalid       = errors.New("bay may only contain letters, numbers and '-'")
	ErrEmailRequired      = errors.New("email is required")
)

var bayPattern = regexp.MustCompile(`^[A-Z0-9-]+$`)

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
	_, err := NormalizeBay(req.Bay)
	return err
}

// NormalizeBay trims and uppercases the bay, and checks it only contains
// letters, digits and '-'. Returns the normalized value on success.
func NormalizeBay(bay string) (string, error) {
	bay = strings.ToUpper(strings.TrimSpace(bay))
	if bay == "" {
		return "", ErrBayRequired
	}
	if !bayPattern.MatchString(bay) {
		return "", ErrBayInvalid
	}
	return bay, nil
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
