package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"

	errorsService "tesina/backend/internal/errors"
	"tesina/backend/internal/models"
	"tesina/backend/internal/repository"
)

type BusService interface {
	JoinBus(ctx context.Context, userID uuid.UUID, req models.JoinBusRequest) error
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

func (s *busService) JoinBus(ctx context.Context, userID uuid.UUID, req models.JoinBusRequest) error {
	if strings.TrimSpace(req.TerminalID) == "" {
		return errorsService.ErrTerminalIDRequired
	}
	terminalUUID, err := uuid.Parse(req.TerminalID)
	if err != nil {
		return errorsService.ErrTerminalIDInvalid
	}
	if strings.TrimSpace(req.Ticket) == "" {
		return errorsService.ErrTicketRequired
	}

	_, err = s.busTerminalRepo.GetByUUID(ctx, terminalUUID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return errorsService.ErrTerminalNotFound
		}
		return fmt.Errorf("busTerminalRepo.GetByUUID: %w", err)
	}

	resp, err := s.busTicketSvc.GetBusTicket(ctx, models.GetBusTicketRequest{TicketString: req.Ticket})
	if err != nil {
		return err
	}
	if resp.StatusCode == http.StatusNotFound {
		return errorsService.ErrTripNotFound
	}

	var ticket models.BusTicket
	if err := json.Unmarshal(resp.Body, &ticket); err != nil {
		return fmt.Errorf("%w: %w", errorsService.ErrUpstreamResponse, err)
	}

	licensePlate := normalizeLicensePlate(ticket.BusLicensePlate)
	groupKey := licensePlate + ":" + terminalUUID.String()

	return s.awaitedTripRepo.Save(ctx, models.AwaitedTrip{
		UserID:   userID,
		GroupKey: groupKey,
	})
}
