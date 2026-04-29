package validators

import (
	"errors"

	"tesina/backend/internal/models"
)

var (
	ErrNoFieldsToUpdate  = errors.New("at least one field must be provided")
	ErrFirstNameEmpty    = errors.New("first_name cannot be empty")
	ErrLastNameEmpty     = errors.New("last_name cannot be empty")
	ErrEmailEmpty        = errors.New("email cannot be empty")
	ErrPasswordEmpty     = errors.New("password cannot be empty")
	ErrDNIEmpty          = errors.New("dni cannot be empty")
)

func ValidateUpdateUserRequest(req models.UpdateUserRequest) error {
	if req.FirstName == nil && req.LastName == nil && req.Email == nil && req.Password == nil && req.DNI == nil {
		return ErrNoFieldsToUpdate
	}
	if req.FirstName != nil && *req.FirstName == "" {
		return ErrFirstNameEmpty
	}
	if req.LastName != nil && *req.LastName == "" {
		return ErrLastNameEmpty
	}
	if req.Email != nil && *req.Email == "" {
		return ErrEmailEmpty
	}
	if req.Password != nil && *req.Password == "" {
		return ErrPasswordEmpty
	}
	if req.DNI != nil && *req.DNI == "" {
		return ErrDNIEmpty
	}
	return nil
}
