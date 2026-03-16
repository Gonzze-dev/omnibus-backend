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

type Platform struct {
	Code          int             `json:"code" gorm:"primaryKey;column:code;autoIncrement"`
	Anden         string          `json:"anden" gorm:"column:anden;not null"`
	Coordinates   json.RawMessage `json:"coordinates" gorm:"column:coordinates;type:jsonb"`
	BusTerminalID uuid.UUID       `json:"bus_terminal_id" gorm:"column:bus_terminal_id;type:uuid;not null"`
}

func (Platform) TableName() string {
	return "platform"
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
