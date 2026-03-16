package service

import "errors"

var (
	ErrTicketStringEmpty = errors.New("ticket string is required")
	ErrUpstreamRequest   = errors.New("failed to request upstream API")
	ErrUpstreamResponse  = errors.New("failed to read upstream response")

	ErrLicensePatentEmpty = errors.New("license_patent is required")
	ErrCodeEmpty          = errors.New("code is required")
	ErrInvalidCode        = errors.New("code must be a valid integer")
	ErrPlatformLookup     = errors.New("failed to look up platform")
	ErrNotification       = errors.New("failed to send notification")
)
