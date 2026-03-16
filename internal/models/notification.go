package models

import "encoding/json"

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
