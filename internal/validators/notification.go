package validators

import (
	"bytes"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"tesina/backend/internal/models"
	"tesina/backend/internal/roles"
)

var (
	ErrLicensePatentEmpty             = errors.New("license_patent is required")
	ErrCodeEmpty                      = errors.New("code is required")
	ErrInvalidCode                    = errors.New("code must be a valid integer")
	ErrNotificationTimeLifeInvalid    = errors.New("time_life must be greater than 0")
	ErrNotificationPayloadEmpty       = errors.New("payload is required")
	ErrNotificationPayloadInvalidJSON = errors.New("payload must be valid JSON")
	ErrNotificationMessageEmpty       = errors.New("payload.message is required")
	ErrNotificationGlobalSuperAdminOnly = errors.New("notification type GLOBAL is only allowed for super_admin")
	ErrNotificationTypeInvalid        = errors.New("notification type must be LOCAL or GLOBAL")
	ErrUnsupportedNotificationRole    = errors.New("unsupported role for notification types")
	ErrBusDelayTypeInvalid            = errors.New("type must be BUS_DELAY")
	ErrBusDelayLicensePatentRequired  = errors.New("license_patent is required")
	ErrBusDelayStartDateRequired      = errors.New("start_date is required")
	ErrBusDelayTimeDelayInvalid       = errors.New("payload.time_delay must be a positive integer")
	ErrCameraNotificationTypeInvalid  = errors.New("type incorrect or missing")
	ErrCodeCameraEmpty                = errors.New("code_camera is required")
	ErrCodeCameraInvalid              = errors.New("code_camera must be a valid integer")
	ErrCameraErrorMessageEmpty        = errors.New("payload.message is required")
	ErrTerminalIDRequired             = errors.New("terminalID query parameter is required")
	ErrInvalidTerminalID              = errors.New("terminalID must be a valid UUID")
	ErrInvalidStartDate               = errors.New("start_date must be in YYYY-MM-DD format")
	ErrInvalidEndDate                 = errors.New("end_date must be in YYYY-MM-DD format")
	ErrEndDateBeforeStart             = errors.New("end_date must be greater than or equal to start_date")
)

// ValidateNotifyPassengersRequest validates required fields and returns the parsed platform code.
func ValidateNotifyPassengersRequest(req models.NotifyPassengersRequest) (int, error) {
	if req.LicensePatent == "" {
		return 0, ErrLicensePatentEmpty
	}
	if req.Code == "" {
		return 0, ErrCodeEmpty
	}
	if req.TimeLife <= 0 {
		return 0, ErrNotificationTimeLifeInvalid
	}
	code, err := strconv.Atoi(req.Code)
	if err != nil {
		return 0, ErrInvalidCode
	}
	return code, nil
}

// ValidateAdminGlobalNotification validates the role and payload for a global notification.
// Returns the parsed time_life on success.
func ValidateAdminGlobalNotification(role string, payloadRaw json.RawMessage) (int, error) {
	if role != roles.SuperAdmin {
		return 0, ErrNotificationGlobalSuperAdminOnly
	}
	trimmed := bytes.TrimSpace(payloadRaw)
	if len(trimmed) == 0 {
		return 0, ErrNotificationPayloadEmpty
	}
	if !json.Valid(trimmed) {
		return 0, ErrNotificationPayloadInvalidJSON
	}
	var tmp struct {
		TimeLife int `json:"time_life"`
	}
	if err := json.Unmarshal(trimmed, &tmp); err != nil || tmp.TimeLife <= 0 {
		return 0, ErrNotificationTimeLifeInvalid
	}
	return tmp.TimeLife, nil
}

// ValidateAdminLocalNotification validates and parses the local notification payload.
func ValidateAdminLocalNotification(payloadRaw json.RawMessage) (models.AdminLocalNotificationPayload, error) {
	if len(bytes.TrimSpace(payloadRaw)) == 0 {
		return models.AdminLocalNotificationPayload{}, ErrNotificationPayloadEmpty
	}
	var p models.AdminLocalNotificationPayload
	if err := json.Unmarshal(payloadRaw, &p); err != nil {
		return models.AdminLocalNotificationPayload{}, ErrNotificationPayloadInvalidJSON
	}
	if p.Message == "" {
		return models.AdminLocalNotificationPayload{}, ErrNotificationMessageEmpty
	}
	if p.TimeLife <= 0 {
		return models.AdminLocalNotificationPayload{}, ErrNotificationTimeLifeInvalid
	}
	return p, nil
}

// ValidateNotifyBusDelayRequest validates all fields of a bus delay notification request.
func ValidateNotifyBusDelayRequest(req models.NotifyBusDelayRequest) error {
	if req.Type != models.PassengerNotificationBUSDelay {
		return ErrBusDelayTypeInvalid
	}
	if strings.TrimSpace(req.LicensePatent) == "" {
		return ErrBusDelayLicensePatentRequired
	}
	if strings.TrimSpace(req.StartDate) == "" {
		return ErrBusDelayStartDateRequired
	}
	if req.Payload.TimeDelay <= 0 {
		return ErrBusDelayTimeDelayInvalid
	}
	if req.Payload.TimeLife <= 0 {
		return ErrNotificationTimeLifeInvalid
	}
	return nil
}

// ValidateCameraErrorRequest validates the camera error request and returns the parsed camera code.
func ValidateCameraErrorRequest(req models.CameraErrorNotifyRequest) (int, error) {
	if req.Type != models.PassengerNotificationCAMERA {
		return 0, ErrCameraNotificationTypeInvalid
	}
	if strings.TrimSpace(req.CodeCamera) == "" {
		return 0, ErrCodeCameraEmpty
	}
	code, err := strconv.Atoi(strings.TrimSpace(req.CodeCamera))
	if err != nil {
		return 0, ErrCodeCameraInvalid
	}
	if strings.TrimSpace(req.Payload.Message) == "" {
		return 0, ErrCameraErrorMessageEmpty
	}
	if req.Payload.TimeLife <= 0 {
		return 0, ErrNotificationTimeLifeInvalid
	}
	return code, nil
}

// ValidateGetNotificationsUserRole validates the TerminalID for the user/passenger role
// and returns the parsed UUID.
func ValidateGetNotificationsUserRole(params models.GetNotificationsParams) (uuid.UUID, error) {
	if params.TerminalID == "" {
		return uuid.Nil, ErrTerminalIDRequired
	}
	tid, err := uuid.Parse(params.TerminalID)
	if err != nil {
		return uuid.Nil, ErrInvalidTerminalID
	}
	return tid, nil
}

// ValidateCommonNotificationParams validates the notification type filter and date range.
func ValidateCommonNotificationParams(params models.GetNotificationsParams) error {
	if params.NotificationType != "" {
		t := models.PassengerNotificationType(params.NotificationType)
		switch t {
		case models.PassengerNotificationBUSArrival,
			models.PassengerNotificationBUSDelay,
			models.PassengerNotificationLocal,
			models.PassengerNotificationGlobal,
			models.PassengerNotificationCAMERA:
		default:
			return ErrNotificationTypeInvalid
		}
	}
	if params.StartDate != "" {
		t, err := time.Parse("2006-01-02", params.StartDate)
		if err != nil {
			return ErrInvalidStartDate
		}
		if params.EndDate != "" {
			t2, err := time.Parse("2006-01-02", params.EndDate)
			if err != nil {
				return ErrInvalidEndDate
			}
			if t2.Before(t) {
				return ErrEndDateBeforeStart
			}
		}
	}
	return nil
}
