package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	errorsService "tesina/backend/internal/errors"
	"tesina/backend/internal/models"
	"tesina/backend/internal/repository"
)

type BusService interface {
	JoinBus(ctx context.Context, userID uuid.UUID, req models.JoinBusRequest) (models.AwaitedTripResponse, error)
	GetAwaitedTrip(ctx context.Context, userID uuid.UUID) (models.AwaitedTripResponse, error)
	LeaveBus(ctx context.Context, userID uuid.UUID) error
}

type busService struct {
	busTerminalRepo repository.BusTerminalRepository
	awaitedTripRepo repository.AwaitedTripRepository
	busTicketSvc    BusTicketService
}

func NewBusService(
	busTerminalRepo repository.BusTerminalRepository,
	awaitedTripRepo repository.AwaitedTripRepository,
	busTicketSvc BusTicketService,
) *busService {
	return &busService{
		busTerminalRepo: busTerminalRepo,
		awaitedTripRepo: awaitedTripRepo,
		busTicketSvc:    busTicketSvc,
	}
}

// JoinBus valida el pasaje contra el sistema de terminales y deja al usuario
// esperando ese colectivo en la terminal elegida. Si ya esperaba otro viaje,
// lo reemplaza.
func (s *busService) JoinBus(ctx context.Context, userID uuid.UUID, req models.JoinBusRequest) (models.AwaitedTripResponse, error) {
	if strings.TrimSpace(req.TerminalID) == "" {
		return models.AwaitedTripResponse{}, errorsService.ErrTerminalIDRequired
	}
	terminalUUID, err := uuid.Parse(strings.TrimSpace(req.TerminalID))
	if err != nil {
		return models.AwaitedTripResponse{}, errorsService.ErrTerminalIDInvalid
	}
	ticketCode := strings.ToUpper(strings.TrimSpace(req.Ticket))
	if ticketCode == "" {
		return models.AwaitedTripResponse{}, errorsService.ErrTicketRequired
	}

	terminal, err := s.getTerminal(ctx, terminalUUID)
	if err != nil {
		return models.AwaitedTripResponse{}, err
	}

	ticket, err := s.fetchTicket(ctx, ticketCode)
	if err != nil {
		return models.AwaitedTripResponse{}, err
	}

	// El pasaje debe pertenecer a la terminal elegida.
	if terminal.ExternalTerminalID == nil ||
		!strings.EqualFold(ticket.TerminalUUID, terminal.ExternalTerminalID.String()) {
		return models.AwaitedTripResponse{}, errorsService.ErrTicketWrongTerminal
	}

	awaited := models.AwaitedTrip{
		UserID:        userID,
		GroupKey:      normalizeLicensePlate(ticket.BusLicensePlate) + ":" + terminal.UUID.String(),
		Ticket:        ticket.Ticket,
		BusTerminalID: terminal.UUID,
		CreatedAt:     time.Now().UTC(),
		NotifiedAt:    nil,
	}
	if err := s.awaitedTripRepo.Upsert(ctx, awaited); err != nil {
		return models.AwaitedTripResponse{}, fmt.Errorf("awaitedTripRepo.Upsert: %w", err)
	}

	return toAwaitedTripResponse(awaited, terminal, ticket), nil
}

// GetAwaitedTrip devuelve el viaje que el usuario está esperando, con los
// datos del pasaje actualizados desde el sistema de terminales.
func (s *busService) GetAwaitedTrip(ctx context.Context, userID uuid.UUID) (models.AwaitedTripResponse, error) {
	awaited, err := s.awaitedTripRepo.GetByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return models.AwaitedTripResponse{}, errorsService.ErrAwaitedTripNotFound
		}
		return models.AwaitedTripResponse{}, fmt.Errorf("awaitedTripRepo.GetByUserID: %w", err)
	}

	terminal, err := s.getTerminal(ctx, awaited.BusTerminalID)
	if err != nil {
		return models.AwaitedTripResponse{}, err
	}

	ticket, err := s.fetchTicket(ctx, awaited.Ticket)
	if err != nil {
		return models.AwaitedTripResponse{}, err
	}

	return toAwaitedTripResponse(awaited, terminal, ticket), nil
}

func (s *busService) LeaveBus(ctx context.Context, userID uuid.UUID) error {
	if err := s.awaitedTripRepo.DeleteByUserID(ctx, userID); err != nil {
		return fmt.Errorf("awaitedTripRepo.DeleteByUserID: %w", err)
	}
	return nil
}

func (s *busService) getTerminal(ctx context.Context, id uuid.UUID) (models.BusTerminal, error) {
	terminal, err := s.busTerminalRepo.GetByUUID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return models.BusTerminal{}, errorsService.ErrTerminalNotFound
		}
		return models.BusTerminal{}, fmt.Errorf("busTerminalRepo.GetByUUID: %w", err)
	}
	return terminal, nil
}

// fetchTicket consulta el pasaje en el backend de terminales.
func (s *busService) fetchTicket(ctx context.Context, ticketCode string) (models.BusTicket, error) {
	resp, err := s.busTicketSvc.GetBusTicket(ctx, models.GetBusTicketRequest{TicketString: ticketCode})
	if err != nil {
		return models.BusTicket{}, err
	}

	switch resp.StatusCode {
	case http.StatusOK:
	case http.StatusNotFound:
		return models.BusTicket{}, errorsService.ErrTripNotFound
	default:
		return models.BusTicket{}, fmt.Errorf("%w: status %d", errorsService.ErrUpstreamResponse, resp.StatusCode)
	}

	var ticket models.BusTicket
	if err := json.Unmarshal(resp.Body, &ticket); err != nil {
		return models.BusTicket{}, fmt.Errorf("%w: %w", errorsService.ErrUpstreamResponse, err)
	}
	return ticket, nil
}

func toAwaitedTripResponse(a models.AwaitedTrip, terminal models.BusTerminal, ticket models.BusTicket) models.AwaitedTripResponse {
	return models.AwaitedTripResponse{
		GroupKey: a.GroupKey,
		Terminal: models.AwaitedTripTerminal{
			UUID: terminal.UUID,
			Name: terminal.Name,
		},
		Trip:       ticket,
		CreatedAt:  a.CreatedAt,
		NotifiedAt: a.NotifiedAt,
	}
}
