package service

import "errors"

var (
	ErrTicketStringEmpty = errors.New("ticket string is required")
	ErrUpstreamRequest   = errors.New("failed to request upstream API")
	ErrUpstreamResponse  = errors.New("failed to read upstream response")

	ErrLicensePatentEmpty = errors.New("license_patent is required")
	ErrCodeEmpty          = errors.New("code is required")
	ErrInvalidCode        = errors.New("code must be a valid integer")
	ErrPlatformLookup          = errors.New("failed to look up platform")
	ErrPlatformMissingTerminal = errors.New("platform is not associated with a bus terminal")
	ErrNotification            = errors.New("failed to send notification")

	ErrEmailAlreadyExists  = errors.New("email already exists")
	ErrDNIAlreadyExists    = errors.New("dni already exists")
	ErrInvalidCredentials  = errors.New("invalid email or password")
	ErrInvalidRefreshToken = errors.New("invalid or expired refresh token")
	ErrUserNotFound        = errors.New("user not found")
	ErrRolNotFound         = errors.New("role not found")
	ErrTerminalNotFound    = errors.New("terminal not found")
	ErrPlatformNotFound    = errors.New("platform not found")
	ErrCityNotFound        = errors.New("city not found")
	ErrTerminalNotOwned    = errors.New("terminal not associated with admin")
	ErrCannotDemoteSelf    = errors.New("cannot demote yourself")
	ErrAlreadyAdmin        = errors.New("user is already admin")
	ErrNotAdmin            = errors.New("user is not admin")
	ErrAlreadySuperAdmin   = errors.New("user is already super admin")
	ErrNotSuperAdmin       = errors.New("user is not super admin")
	ErrMissingFields       = errors.New("missing required fields")
	ErrCityAlreadyExists   = errors.New("city with this postal code already exists")
)
