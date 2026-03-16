package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"tesina/backend/internal/models"
)

type PasajeService interface {
	GetPasaje(ctx context.Context, req models.GetPasajeRequest) (models.PasajeResponse, error)
}

type pasajeService struct {
	httpClient  *http.Client
	upstreamURL string
}

func NewPasajeService(httpClient *http.Client, upstreamURL string) *pasajeService {
	return &pasajeService{
		httpClient:  httpClient,
		upstreamURL: upstreamURL,
	}
}

func (s *pasajeService) GetPasaje(ctx context.Context, req models.GetPasajeRequest) (models.PasajeResponse, error) {
	if req.TicketString == "" {
		return models.PasajeResponse{}, ErrTicketStringEmpty
	}

	url := fmt.Sprintf("%s/pasajes/%s", s.upstreamURL, req.TicketString)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return models.PasajeResponse{}, fmt.Errorf("%w: %w", ErrUpstreamRequest, err)
	}

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return models.PasajeResponse{}, fmt.Errorf("%w: %w", ErrUpstreamRequest, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return models.PasajeResponse{}, fmt.Errorf("%w: %w", ErrUpstreamResponse, err)
	}

	if resp.StatusCode == http.StatusOK {
		normalized, err := normalizePasaje(body)
		if err == nil {
			body = normalized
		}
	}

	return models.PasajeResponse{
		Body:       body,
		StatusCode: resp.StatusCode,
	}, nil
}

func normalizePasaje(raw []byte) ([]byte, error) {
	var p models.Pasaje
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
