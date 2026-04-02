package models

import "encoding/json"

type GetBusTicketRequest struct {
	TicketString string `json:"ticket_string"`
}

type BusTicketResponse struct {
	Body       json.RawMessage `json:"body"`
	StatusCode int             `json:"status_code"`
}

type TripCity struct {
	CityName  string `json:"city_name"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
	Order     int    `json:"order"`
}

type BusTicket struct {
	PostalCode      string     `json:"postal_code"`
	BusTerminalName string     `json:"bus_terminal_name"`
	Ticket          string     `json:"ticket"`
	DNI             string     `json:"dni"`
	Name            string     `json:"name"`
	BusLicensePlate string     `json:"bus_license_plate"`
	Enterprise      string     `json:"enterprise"`
	StartDate       string     `json:"start_date"`
	EndDate         string     `json:"end_date"`
	TripCity        []TripCity `json:"trip_city"`
	UUID            string     `json:"uuid"`
}
