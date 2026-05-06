package errors

import "errors"

var (
	ErrTicketStringEmpty = errors.New("ticket string is required")
	ErrUpstreamRequest   = errors.New("failed to request upstream API")
	ErrUpstreamResponse  = errors.New("failed to read upstream response")

	ErrPlatformLookup          = errors.New("failed to look up platform")
	ErrPlatformMissingTerminal = errors.New("platform is not associated with a bus terminal")
	ErrNotification            = errors.New("failed to send notification")

	ErrEmailAlreadyExists        = errors.New("email already exists")
	ErrInvalidCredentials        = errors.New("invalid email or password")
	ErrInvalidRefreshToken       = errors.New("invalid or expired refresh token")
	ErrInvalidPasswordResetToken = errors.New("invalid or expired password reset token")
	ErrUserNotFound              = errors.New("user not found")
	ErrRolNotFound               = errors.New("role not found")
	ErrTerminalNotFound          = errors.New("terminal not found")
	ErrPlatformNotFound          = errors.New("platform not found")
	ErrCityNotFound              = errors.New("city not found")
	ErrTerminalNotOwned          = errors.New("terminal not associated with admin")
	ErrCannotDemoteSelf          = errors.New("cannot demote yourself")
	ErrAlreadyAdmin              = errors.New("user is already admin")
	ErrNotAdmin                  = errors.New("user is not admin")
	ErrAlreadySuperAdmin         = errors.New("user is already super admin")
	ErrNotSuperAdmin             = errors.New("user is not super admin")
	ErrMissingFields             = errors.New("missing required fields")
	ErrCityAlreadyExists         = errors.New("city with this postal code already exists")

	ErrTerminalUUIDRequired           = errors.New("query terminaluuid is required for super_admin")
	ErrTerminalUUIDRequiredMultiAdmin = errors.New("query terminaluuid is required when admin has more than one terminal")
	ErrAdminNoTerminal                = errors.New("admin has no associated terminal")
	ErrInvalidTerminalUUID            = errors.New("invalid terminaluuid")

	ErrExternalTerminalIDRequired    = errors.New("external_terminal_id is required")
	ErrInvalidExternalTerminalID     = errors.New("invalid external_terminal_id")
	ErrExternalTerminalIDAlreadyUsed = errors.New("external_terminal_id is already registered")

	ErrBusDelayTerminalUUIDRequired  = errors.New("uuid_terminal is required")
	ErrExternalTerminalNotConfigured = errors.New("terminal has no external_terminal_id configured")
	ErrTripNotRegistered             = errors.New("trip is not registered in the terminal system; notification cannot be sent")

	ErrNotificationNotFound        = errors.New("notification not found")
	ErrNotificationDeleteForbidden = errors.New("you do not have permission to delete this notification")
	ErrUserCannotDeleteNotification = errors.New("users cannot delete notifications")

	ErrTripNotFound       = errors.New("trip not found")
	ErrTerminalIDRequired = errors.New("terminalId is required")
	ErrTerminalIDInvalid  = errors.New("terminalId must be a valid UUID")
	ErrTicketRequired     = errors.New("ticket is required")
)
