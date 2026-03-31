package models

import "encoding/json"

// PassengerNotificationType categorizes realtime messages to passengers.
type PassengerNotificationType string

const (
	PassengerNotificationBUSArrival PassengerNotificationType = "BUS_ARRIVAL"
	PassengerNotificationBUSDelay   PassengerNotificationType = "BUS_DELAY"
	PassengerNotificationLocal      PassengerNotificationType = "LOCAL"
	PassengerNotificationGlobal     PassengerNotificationType = "GLOBAL"
)

// PassengerNotificationMessage is sent to the frontend (SendToFrontend).
type PassengerNotificationMessage struct {
	Type    PassengerNotificationType `json:"type"`
	Payload json.RawMessage           `json:"payload"`
}

type NotifyPassengersRequest struct {
	LicensePatent string `json:"license_patent"`
	Code          string `json:"code"`
}

type NotifyPassengersResponse struct {
	Message string `json:"message"`
}

type PlatformInfo struct {
	Anden       string          `json:"anden"`
	Coordinates json.RawMessage `json:"coordinates"`
}

// AdminLocalNotificationPayload is the expected shape of payload when type is LOCAL.
type AdminLocalNotificationPayload struct {
	Message string `json:"message"`
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
	TimeDelay int `json:"time_delay"`
}

type NotifyBusDelayResponse struct {
	Message string `json:"message"`
}
