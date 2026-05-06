package models

import "github.com/google/uuid"

type AwaitedTrip struct {
	UserID   uuid.UUID `gorm:"primaryKey;column:user_id;type:uuid"`
	GroupKey string    `gorm:"column:group_key;not null"`
}

func (AwaitedTrip) TableName() string {
	return "awaited_trip"
}

type JoinBusRequest struct {
	TerminalID string `json:"terminalId"`
	Ticket     string `json:"ticket"`
}
