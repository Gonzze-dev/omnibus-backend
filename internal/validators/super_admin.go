package validators

import (
	"errors"

	"github.com/google/uuid"

	"tesina/backend/internal/models"
)

var (
	ErrTerminalNameRequired       = errors.New("name is required")
	ErrTerminalPostalCodeRequired = errors.New("postal_code is required")
	ErrExternalTerminalIDRequired = errors.New("external_terminal_id is required")
)

func ValidateCreateBusTerminalRequest(req models.CreateBusTerminalRequest) error {
	if req.PostalCode == "" {
		return ErrTerminalPostalCodeRequired
	}
	if req.Name == "" {
		return ErrTerminalNameRequired
	}
	if req.ExternalTerminalID == uuid.Nil {
		return ErrExternalTerminalIDRequired
	}
	return nil
}

func ValidateUpdateBusTerminalRequest(req models.UpdateBusTerminalRequest) error {
	if req.ExternalTerminalID != nil && *req.ExternalTerminalID == uuid.Nil {
		return ErrExternalTerminalIDRequired
	}
	return nil
}

func ValidatePromoteSuperRequest(req models.PromoteSuperRequest) error {
	if req.Email == "" {
		return ErrEmailRequired
	}
	return nil
}

func ValidateDemoteSuperRequest(req models.DemoteSuperRequest) error {
	if req.Email == "" {
		return ErrEmailRequired
	}
	return nil
}
