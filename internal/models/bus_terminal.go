package models

import (
	"encoding/json"

	"github.com/google/uuid"
)

type BusTerminal struct {
	UUID       uuid.UUID  `json:"uuid" gorm:"primaryKey;column:uuid;type:uuid;default:uuid_generate_v4()"`
	PostalCode string     `json:"postal_code" gorm:"column:postal_code;not null"`
	Name       string     `json:"name" gorm:"column:name;not null"`
	City       *City      `json:"city,omitempty" gorm:"foreignKey:PostalCode;references:PostalCode"`
	Platforms  []Platform `json:"platforms,omitempty" gorm:"foreignKey:BusTerminalID;references:UUID"`
}

func (BusTerminal) TableName() string {
	return "bus_terminal"
}

type Platform struct {
	Code          int             `json:"code" gorm:"primaryKey;column:code;autoIncrement"`
	Anden         string          `json:"anden" gorm:"column:anden;not null"`
	Coordinates   json.RawMessage `json:"coordinates" gorm:"column:coordinates;type:jsonb"`
	BusTerminalID uuid.UUID       `json:"bus_terminal_id" gorm:"column:bus_terminal_id;type:uuid;not null"`
	BusTerminal   *BusTerminal    `json:"bus_terminal,omitempty" gorm:"foreignKey:BusTerminalID;references:UUID"`
}

func (Platform) TableName() string {
	return "platform"
}

type PlatformResponse struct {
	Code        int             `json:"code"`
	Anden       string          `json:"anden"`
	Coordinates json.RawMessage `json:"coordinates"`
}

type BusTerminalWithPlatformsResponse struct {
	UUID       uuid.UUID          `json:"uuid"`
	PostalCode string             `json:"postal_code"`
	Name       string             `json:"name"`
	Platforms  []PlatformResponse `json:"platforms"`
}

func ToBusTerminalWithPlatformsResponse(terminals []BusTerminal) []BusTerminalWithPlatformsResponse {
	resp := make([]BusTerminalWithPlatformsResponse, len(terminals))
	for i, t := range terminals {
		platforms := make([]PlatformResponse, len(t.Platforms))
		for j, p := range t.Platforms {
			platforms[j] = PlatformResponse{
				Code:        p.Code,
				Anden:       p.Anden,
				Coordinates: p.Coordinates,
			}
		}
		resp[i] = BusTerminalWithPlatformsResponse{
			UUID:       t.UUID,
			PostalCode: t.PostalCode,
			Name:       t.Name,
			Platforms:  platforms,
		}
	}
	return resp
}

type BusTerminalResponse struct {
	UUID       uuid.UUID `json:"uuid"`
	PostalCode string    `json:"postal_code"`
	Name       string    `json:"name"`
}

func ToBusTerminalResponses(terminals []BusTerminal) []BusTerminalResponse {
	resp := make([]BusTerminalResponse, len(terminals))
	for i, t := range terminals {
		resp[i] = BusTerminalResponse{
			UUID:       t.UUID,
			PostalCode: t.PostalCode,
			Name:       t.Name,
		}
	}
	return resp
}

type CreateBusTerminalRequest struct {
	PostalCode string `json:"postal_code"`
	Name       string `json:"name"`
}

type CreatePlatformRequest struct {
	Anden         string          `json:"anden"`
	Coordinates   json.RawMessage `json:"coordinates"`
	BusTerminalID uuid.UUID       `json:"bus_terminal_id"`
}

type UpdateBusTerminalRequest struct {
	PostalCode *string `json:"postal_code,omitempty"`
	Name       *string `json:"name,omitempty"`
}

type UpdatePlatformRequest struct {
	Anden       *string          `json:"anden,omitempty"`
	Coordinates *json.RawMessage `json:"coordinates,omitempty"`
}

type PromoteAdminRequest struct {
	Email         string    `json:"email"`
	BusTerminalID uuid.UUID `json:"bus_terminal_id"`
}

type DemoteAdminRequest struct {
	Email         string    `json:"email"`
	BusTerminalID uuid.UUID `json:"bus_terminal_id"`
}

type PromoteSuperRequest struct {
	Email string `json:"email"`
}

type DemoteSuperRequest struct {
	Email string `json:"email"`
}
