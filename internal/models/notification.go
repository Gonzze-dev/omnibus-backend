package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Notification struct {
	ID         uuid.UUID       `json:"id" gorm:"primaryKey;column:id;type:uuid;default:uuid_generate_v4()"`
	GroupKey   *string         `json:"group_key,omitempty" gorm:"column:group_key"`
	GroupName  string          `json:"group_name" gorm:"column:group_name;not null"`
	Expiration time.Time       `json:"expiration" gorm:"column:expiration;not null"`
	Date       time.Time       `json:"date" gorm:"column:date;not null"`
	Payload    json.RawMessage `json:"payload" gorm:"column:payload;type:jsonb;not null"`
}

func (Notification) TableName() string {
	return "notifications"
}

// PassengerNotificationType categorizes realtime messages to passengers.
type PassengerNotificationType string

const (
	PassengerNotificationBUSArrival PassengerNotificationType = "BUS_ARRIVAL"
	PassengerNotificationBUSDelay   PassengerNotificationType = "BUS_DELAY"
	PassengerNotificationLocal      PassengerNotificationType = "LOCAL"
	PassengerNotificationGlobal     PassengerNotificationType = "GLOBAL"
	PassengerNotificationCAMERA     PassengerNotificationType = "CAMERA"
)

// PassengerNotificationMessage is sent to the frontend (SendToFrontend).
type PassengerNotificationMessage struct {
	Type    PassengerNotificationType `json:"type"`
	Payload json.RawMessage           `json:"payload"`
}

type NotifyPassengersRequest struct {
	LicensePatent string `json:"license_patent"`
	Code          string `json:"code"`
	TimeLife      int    `json:"time_life"`
}

type NotifyPassengersResponse struct {
	Message string `json:"message"`
}

type PlatformInfo struct {
	ID          uuid.UUID       `json:"id"`
	Anden       string          `json:"anden"`
	Coordinates json.RawMessage `json:"coordinates"`
	TimeLife    int             `json:"time_life"`
}

// AdminLocalNotificationPayload is the expected shape of payload when type is LOCAL.
type AdminLocalNotificationPayload struct {
	ID       string `json:"id"`
	Message  string `json:"message"`
	TimeLife int    `json:"time_life"`
}

// AdminSendNotificationRequest is the JSON body for POST /api/admin/notifications.
// type LOCAL: payload must be {"message":"..."}; query terminaluuid rules apply (see service).
// type GLOBAL: super_admin only; payload is any JSON value; SignalR SendToFrontendGlobal sends {type, payload} like LOCAL.
type AdminSendNotificationRequest struct {
	Type    PassengerNotificationType `json:"type"`
	Payload json.RawMessage           `json:"payload"`
}

type AdminSendNotificationResponse struct {
	Message string `json:"message"`
}

// NotifyBusDelayRequest is the JSON body for POST /api/admin/notify-bus-delay.
// Admin: uuid_terminal optional if exactly one terminal is assigned; required if several.
// Super admin: uuid_terminal is required (target terminal).
type NotifyBusDelayRequest struct {
	Type          PassengerNotificationType `json:"type"`
	LicensePatent string                    `json:"license_patent"`
	UUIDTerminal  string                    `json:"uuid_terminal"`
	StartDate     string                    `json:"start_date"`
	Payload       NotifyBusDelayPayload     `json:"payload"`
}

type NotifyBusDelayPayload struct {
	ID            string `json:"id"`
	LicensePatent string `json:"license_patent"`
	TimeDelay     int    `json:"time_delay"`
	TimeLife      int    `json:"time_life"`
}

type NotifyBusDelayResponse struct {
	Message string `json:"message"`
}

type NotifyDelayBusKeys struct {
	Key           string `json:"key"`
	LicensePatent string `json:"licensePatent"`
	TerminalID    string `json:"terminalId"`
}

// AdminNotificationTypesResponse is the JSON body for GET /api/admin/notification-types.
// Types depend on the caller role (admin vs super_admin).
type AdminNotificationTypesResponse struct {
	Types []PassengerNotificationType `json:"types"`
}

// CameraErrorNotifyRequest is the JSON body for POST /notify_camera_error (X-API-Key).
// code_camera is the platform code; it resolves to the bus terminal UUID for SignalR grouping.
type CameraErrorNotifyRequest struct {
	Type       PassengerNotificationType `json:"type"`
	CodeCamera string                  `json:"code_camera"`
	Payload    CameraErrorNotifyPayload  `json:"payload"`
}

type CameraErrorNotifyPayload struct {
	ID       string `json:"id"`
	Message  string `json:"message"`
	TimeLife int    `json:"time_life"`
}

// CameraErrorNotifyResponse echoes the notification sent to admins via NotifyAdminFromCamera.
type CameraErrorNotifyResponse struct {
	Type    PassengerNotificationType `json:"type"`
	Payload CameraErrorNotifyPayload  `json:"payload"`
}

// GetNotificationsParams carries raw query params from the handler to the service.
type GetNotificationsParams struct {
	TerminalID       string // UUID string; required for user role
	NotificationType string // optional: one of the PassengerNotificationType constants
	ExpirationFilter string // "true" = only expired; "" / "false" = only non-expired (default)
	LicensePlate     string // optional: raw license plate
	StartDate        string // optional: "YYYY-MM-DD"
	EndDate          string // optional: "YYYY-MM-DD", ignored if StartDate is empty
	Limit            int    // default 10
	Offset           int    // default 0
}

// NotificationFilters is built by the service from RBAC logic + query params
// and forwarded to the repository's ListWithFilters method.
type NotificationFilters struct {
	GroupKeyIsNull     bool
	GroupKeyExact      []string // WHERE group_key IN (values...)
	GroupKeyLike       []string // WHERE group_key LIKE pattern (one per entry)
	ExcludeAdminGroups bool     // WHERE group_name NOT ILIKE '%admin%'

	NotificationType *PassengerNotificationType
	OnlyExpired      *bool // nil → default: expiration > NOW()
	StartDate        *time.Time
	EndDate          *time.Time // only applied when StartDate != nil

	Limit  int
	Offset int
}

// NotificationResponseItem is a single notification in the paginated response.
type NotificationResponseItem struct {
	ID         uuid.UUID       `json:"id"`
	Expiration string          `json:"expiration"` // "2006-01-02 15:04:05"
	Date       string          `json:"date"`        // "2006-01-02"
	Data       json.RawMessage `json:"data"`        // DB payload column with inner id stripped
}

// GetNotificationsResponse is the paginated response for GET /api/notifications.
type GetNotificationsResponse struct {
	TotalPages    int                        `json:"total_pages"`
	NumberPage    int                        `json:"number_page"`
	Notifications []NotificationResponseItem `json:"notifications"`
}
