package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/google/uuid"

	"tesina/backend/internal/models"
)

type BusTicketService interface {
	GetBusTicket(ctx context.Context, req models.GetBusTicketRequest) (models.BusTicketResponse, error)
	ExternalTerminalExists(ctx context.Context, externalTerminalUUID uuid.UUID) (bool, error)
	TripExists(ctx context.Context, externalTerminalUUID uuid.UUID, startDate, licensePlate string) (bool, error)
}

type busTicketService struct {
	httpClient  *http.Client
	upstreamURL string
}

func NewBusTicketService(httpClient *http.Client, upstreamURL string) *busTicketService {
	return &busTicketService{
		httpClient:  httpClient,
		upstreamURL: upstreamURL,
	}
}

func (s *busTicketService) GetBusTicket(ctx context.Context, req models.GetBusTicketRequest) (models.BusTicketResponse, error) {
	if req.TicketString == "" {
		return models.BusTicketResponse{}, ErrTicketStringEmpty
	}

	url := fmt.Sprintf("%s/bus_tickets/%s", s.upstreamURL, req.TicketString)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return models.BusTicketResponse{}, fmt.Errorf("%w: %w", ErrUpstreamRequest, err)
	}

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return models.BusTicketResponse{}, fmt.Errorf("%w: %w", ErrUpstreamRequest, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return models.BusTicketResponse{}, fmt.Errorf("%w: %w", ErrUpstreamResponse, err)
	}

	if resp.StatusCode == http.StatusOK {
		normalized, err := normalizeBusTicket(body)
		if err == nil {
			body = normalized
		}
	}

	return models.BusTicketResponse{
		Body:       body,
		StatusCode: resp.StatusCode,
	}, nil
}

func normalizeBusTicket(raw []byte) ([]byte, error) {
	var p models.BusTicket
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrUpstreamResponse, err)
	}

	p.BusTerminalName = strings.ToUpper(p.BusTerminalName)
	p.Ticket = strings.ToUpper(p.Ticket)
	p.Name = strings.ToUpper(p.Name)
	p.BusLicensePlate = strings.ToUpper(strings.ReplaceAll(p.BusLicensePlate, " ", ""))
	p.Enterprise = strings.ToUpper(p.Enterprise)

	for i := range p.TripCity {
		p.TripCity[i].CityName = strings.ToUpper(p.TripCity[i].CityName)
	}

	return json.Marshal(p)
}

func (s *busTicketService) ExternalTerminalExists(ctx context.Context, externalTerminalUUID uuid.UUID) (bool, error) {
	if externalTerminalUUID == uuid.Nil {
		return false, ErrExternalTerminalIDRequired
	}

	u, err := url.Parse(s.upstreamURL)
	if err != nil {
		return false, fmt.Errorf("%w: %w", ErrUpstreamRequest, err)
	}
	u = u.JoinPath("terminal", "exist")
	q := u.Query()
	q.Set("uuid", externalTerminalUUID.String())
	u.RawQuery = q.Encode()

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return false, fmt.Errorf("%w: %w", ErrUpstreamRequest, err)
	}

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return false, fmt.Errorf("%w: %w", ErrUpstreamRequest, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("%w: %w", ErrUpstreamResponse, err)
	}

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("%w: status %d", ErrUpstreamResponse, resp.StatusCode)
	}

	return parseUpstreamBoolBody(body)
}

func (s *busTicketService) TripExists(ctx context.Context, externalTerminalUUID uuid.UUID, startDate, licensePlate string) (bool, error) {
	if externalTerminalUUID == uuid.Nil {
		return false, ErrExternalTerminalIDRequired
	}
	if strings.TrimSpace(startDate) == "" {
		return false, ErrBusDelayStartDateRequired
	}
	if strings.TrimSpace(licensePlate) == "" {
		return false, ErrBusDelayLicensePatentRequired
	}

	u, err := url.Parse(s.upstreamURL)
	if err != nil {
		return false, fmt.Errorf("%w: %w", ErrUpstreamRequest, err)
	}
	u = u.JoinPath("terminal", "trip", "exist")
	q := u.Query()
	q.Set("uuid", externalTerminalUUID.String())
	q.Set("start_date", strings.TrimSpace(startDate))
	q.Set("license_plate", licensePlate)
	u.RawQuery = q.Encode()

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return false, fmt.Errorf("%w: %w", ErrUpstreamRequest, err)
	}

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return false, fmt.Errorf("%w: %w", ErrUpstreamRequest, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("%w: %w", ErrUpstreamResponse, err)
	}

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("%w: status %d", ErrUpstreamResponse, resp.StatusCode)
	}

	return parseUpstreamBoolBody(body)
}

func normalizeLicensePlate(s string) string {
	return strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(s), " ", ""))
}

func parseUpstreamBoolBody(body []byte) (bool, error) {
	trimmed := bytes.TrimSpace(body)
	var b bool
	if err := json.Unmarshal(trimmed, &b); err == nil {
		return b, nil
	}

	var obj struct {
		Exist *bool `json:"exist"`
	}
	if err := json.Unmarshal(trimmed, &obj); err == nil && obj.Exist != nil {
		return *obj.Exist, nil
	}

	switch strings.ToLower(string(trimmed)) {
	case "true":
		return true, nil
	case "false":
		return false, nil
	default:
		return false, fmt.Errorf("%w: unexpected body %q", ErrUpstreamResponse, trimmed)
	}
}
