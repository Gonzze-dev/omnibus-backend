package models

import (
	"encoding/json"

	"github.com/google/uuid"
)

type BusTerminal struct {
	UUID                uuid.UUID  `json:"uuid" gorm:"primaryKey;column:uuid;type:uuid;default:uuid_generate_v4()"`
	ExternalTerminalID  *uuid.UUID `json:"external_terminal_id,omitempty" gorm:"column:external_terminal_id;type:uuid"`
	PostalCode          string     `json:"postal_code" gorm:"column:postal_code;not null"`
	Name                string     `json:"name" gorm:"column:name;not null"`
	City       *City      `json:"city,omitempty" gorm:"foreignKey:PostalCode;references:PostalCode"`
	Platforms  []Platform `json:"platforms,omitempty" gorm:"foreignKey:BusTerminalID;references:UUID"`
}

func (BusTerminal) TableName() string {
	return "bus_terminal"
}

type Platform struct {
	Code          int             `json:"code" gorm:"primaryKey;column:code;autoIncrement"`
	Bay         string          `json:"bay" gorm:"column:bay;not null"`
	Coordinates   json.RawMessage `json:"coordinates" gorm:"column:coordinates;type:jsonb"`
	BusTerminalID uuid.UUID       `json:"bus_terminal_id" gorm:"column:bus_terminal_id;type:uuid;not null"`
	BusTerminal   *BusTerminal    `json:"bus_terminal,omitempty" gorm:"foreignKey:BusTerminalID;references:UUID"`
}

func (Platform) TableName() string {
	return "platform"
}

type PlatformResponse struct {
	Code        int             `json:"code"`
	Bay       string          `json:"bay"`
	Coordinates json.RawMessage `json:"coordinates"`
}

type BusTerminalWithPlatformsResponse struct {
	UUID               uuid.UUID          `json:"uuid"`
	ExternalTerminalID *uuid.UUID         `json:"external_terminal_id,omitempty"`
	PostalCode         string             `json:"postal_code"`
	Name               string             `json:"name"`
	Platforms          []PlatformResponse `json:"platforms"`
}

func ToBusTerminalWithPlatformsResponse(terminals []BusTerminal) []BusTerminalWithPlatformsResponse {
	resp := make([]BusTerminalWithPlatformsResponse, len(terminals))
	for i, t := range terminals {
		platforms := make([]PlatformResponse, len(t.Platforms))
		for j, p := range t.Platforms {
			platforms[j] = PlatformResponse{
				Code:        p.Code,
				Bay:       p.Bay,
				Coordinates: p.Coordinates,
			}
		}
		resp[i] = BusTerminalWithPlatformsResponse{
			UUID:               t.UUID,
			ExternalTerminalID: t.ExternalTerminalID,
			PostalCode:         t.PostalCode,
			Name:               t.Name,
			Platforms:          platforms,
		}
	}
	return resp
}

type BusTerminalResponse struct {
	UUID               uuid.UUID  `json:"uuid"`
	ExternalTerminalID *uuid.UUID `json:"external_terminal_id,omitempty"`
	PostalCode         string     `json:"postal_code"`
	Name               string     `json:"name"`
}

func ToBusTerminalResponses(terminals []BusTerminal) []BusTerminalResponse {
	resp := make([]BusTerminalResponse, len(terminals))
	for i, t := range terminals {
		resp[i] = BusTerminalResponse{
			UUID:               t.UUID,
			ExternalTerminalID: t.ExternalTerminalID,
			PostalCode:         t.PostalCode,
			Name:               t.Name,
		}
	}
	return resp
}

type CreateBusTerminalRequest struct {
	ExternalTerminalID uuid.UUID `json:"external_terminal_id"`
	PostalCode         string    `json:"postal_code"`
	Name               string    `json:"name"`
}

type CreatePlatformRequest struct {
	Bay         string          `json:"bay"`
	Coordinates   json.RawMessage `json:"coordinates"`
	BusTerminalID uuid.UUID       `json:"bus_terminal_id"`
}

type UpdateBusTerminalRequest struct {
	ExternalTerminalID *uuid.UUID `json:"external_terminal_id,omitempty"`
	PostalCode         *string    `json:"postal_code,omitempty"`
	Name               *string    `json:"name,omitempty"`
}

type UpdatePlatformRequest struct {
	Bay       *string          `json:"bay,omitempty"`
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

// ListTerminalsParams are the pagination parameters of GET /api/super/terminals.
type ListTerminalsParams struct {
	Page  int    // default 1
	Limit int    // default 10
	Order string // "ASC" or "DESC", default "DESC"
}

// ListTerminalsResponse is the paginated terminals payload.
type ListTerminalsResponse struct {
	Terminals     []BusTerminal `json:"terminals"`
	Page          int           `json:"page"`
	Next          *int          `json:"next"`
	Prev          *int          `json:"prev"`
	Elements      int           `json:"elements"`
	TotalElements int64         `json:"total_elements"`
}

// CountTerminalsResponse is the payload of GET /api/super/terminals/count.
type CountTerminalsResponse struct {
	Total int64 `json:"total"`
}

// ListPlatformsParams are the parameters of GET /api/admin/platforms.
type ListPlatformsParams struct {
	Page  int    // default 1
	Limit int    // default 10
	Order string // "ASC" or "DESC", default "DESC"
}

// ListPlatformsResponse is the paginated payload of platforms grouped by terminal.
type ListPlatformsResponse struct {
	Platforms     []BusTerminalWithPlatformsResponse `json:"platforms"`
	Page          int                                `json:"page"`
	Next          *int                               `json:"next"`
	Prev          *int                               `json:"prev"`
	Elements      int                                `json:"elements"`
	TotalElements int64                              `json:"total_elements"`
}

// CountPlatformsResponse is the payload of GET /api/admin/platforms/count.
type CountPlatformsResponse struct {
	Total int64 `json:"total"`
}
