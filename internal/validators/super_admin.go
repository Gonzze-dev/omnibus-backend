package validators

import (
	"errors"
	"strings"
	"unicode"

	"github.com/google/uuid"

	"tesina/backend/internal/models"
)

var (
	ErrTerminalNameRequired       = errors.New("name is required")
	ErrTerminalPostalCodeRequired = errors.New("postal_code is required")
	ErrExternalTerminalIDRequired = errors.New("external_terminal_id is required")
	ErrTerminalNameInvalid        = errors.New("name must contain only letters and numbers")
)

// NormalizeTerminalName recorta espacios y pasa el nombre a mayúsculas.
func NormalizeTerminalName(name string) string {
	return strings.ToUpper(strings.TrimSpace(name))
}

// validateTerminalName acepta únicamente letras y números (se permiten espacios internos como separador).
func validateTerminalName(name string) error {
	if name == "" {
		return ErrTerminalNameRequired
	}
	for _, r := range name {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != ' ' {
			return ErrTerminalNameInvalid
		}
	}
	return nil
}

func ValidateCreateBusTerminalRequest(req models.CreateBusTerminalRequest) error {
	if req.PostalCode == "" {
		return ErrTerminalPostalCodeRequired
	}
	if err := validateTerminalName(req.Name); err != nil {
		return err
	}
	if req.ExternalTerminalID == uuid.Nil {
		return ErrExternalTerminalIDRequired
	}
	return nil
}

func ValidateUpdateBusTerminalRequest(req models.UpdateBusTerminalRequest) error {
	if req.Name != nil {
		if err := validateTerminalName(*req.Name); err != nil {
			return err
		}
	}
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
