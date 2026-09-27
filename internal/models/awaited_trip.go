package models

import (
	"time"

	"github.com/google/uuid"
)

// AwaitedTrip es el viaje que un pasajero está esperando en una terminal.
// Un usuario espera un único viaje a la vez (user_id es la PK).
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

// AwaitedTripResponse es lo que necesita el frontend para quedar a la espera:
// los datos del pasaje y el group_key para unirse al canal de realtime.
type AwaitedTripResponse struct {
	GroupKey   string              `json:"group_key"`
	Terminal   AwaitedTripTerminal `json:"terminal"`
	Trip       BusTicket           `json:"trip"`
	CreatedAt  time.Time           `json:"created_at"`
	NotifiedAt *time.Time          `json:"notified_at"`
}
