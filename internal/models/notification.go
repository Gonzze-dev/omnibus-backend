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
