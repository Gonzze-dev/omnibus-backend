package models

import "github.com/google/uuid"

type UserTerminal struct {
	UserID        uuid.UUID `json:"user_id" gorm:"primaryKey;column:user_id;type:uuid"`
	BusTerminalID uuid.UUID `json:"bus_terminal_id" gorm:"primaryKey;column:bus_terminal_id;type:uuid"`
}

func (UserTerminal) TableName() string {
	return "user_terminal"
}
