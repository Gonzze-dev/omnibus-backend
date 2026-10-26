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
)

func ValidateUpdateUserRequest(req models.UpdateUserRequest) error {
	if req.FirstName == nil && req.LastName == nil && req.Email == nil && req.Password == nil {
		return ErrNoFieldsToUpdate
	}
	if req.FirstName != nil && *req.FirstName == "" {
		return ErrFirstNameEmpty
	}
	if req.LastName != nil && *req.LastName == "" {
		return ErrLastNameEmpty
	}
	if req.Email != nil {
		if *req.Email == "" {
			return ErrEmailEmpty
		}
		if err := validateEmailFormat(*req.Email); err != nil {
			return err
		}
	}
	if req.Password != nil && *req.Password == "" {
		return ErrPasswordEmpty
	}
	return nil
}
