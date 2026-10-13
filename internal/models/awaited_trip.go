package models

import (
	"time"

	"github.com/google/uuid"
)

// AwaitedTrip is the trip a passenger is waiting for at a terminal.
// A user waits for a single trip at a time (user_id is the PK).
type AwaitedTrip struct {
	UserID        uuid.UUID  `gorm:"primaryKey;column:user_id;type:uuid"`
	GroupKey      string     `gorm:"column:group_key;not null"`
	Ticket        string     `gorm:"column:ticket;not null"`
	BusTerminalID uuid.UUID  `gorm:"column:bus_terminal_id;type:uuid;not null"`
	CreatedAt     time.Time  `gorm:"column:created_at;not null"`
	NotifiedAt    *time.Time `gorm:"column:notified_at"`
}

func (AwaitedTrip) TableName() string {
	return "awaited_trip"
}

type JoinBusRequest struct {
	TerminalID string `json:"terminalId"`
	Ticket     string `json:"ticket"`
}

type AwaitedTripTerminal struct {
	UUID uuid.UUID `json:"uuid"`
	Name string    `json:"name"`
}

// AwaitedTripResponse is what the frontend needs to start waiting:
// the bus ticket data and the group_key to join the realtime channel.
type AwaitedTripResponse struct {
	GroupKey   string              `json:"group_key"`
	Terminal   AwaitedTripTerminal `json:"terminal"`
	Trip       BusTicket           `json:"trip"`
	CreatedAt  time.Time           `json:"created_at"`
	NotifiedAt *time.Time          `json:"notified_at"`
}
