package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/google/uuid"

	"tesina/backend/internal/models"
	"tesina/backend/internal/repository"
)

type RealtimeNotifier interface {
	Invoke(ctx context.Context, method string, args ...any) error
}

type NotificationService interface {
	NotifyPassengers(ctx context.Context, req models.NotifyPassengersRequest) (models.NotifyPassengersResponse, error)
	SendAdminNotification(ctx context.Context, userID uuid.UUID, role string, queryTerminalUUID string, req models.AdminSendNotificationRequest) (models.AdminSendNotificationResponse, error)
	NotifyBusDelay(ctx context.Context, userID uuid.UUID, role string, req models.NotifyBusDelayRequest) (models.NotifyBusDelayResponse, error)
	ListAdminSelectableNotificationTypes(ctx context.Context, role string) (models.AdminNotificationTypesResponse, error)
}

type notificationService struct {
	platformRepo     repository.PlatformRepository
	userTerminalRepo repository.UserTerminalRepository
	busTerminalRepo  repository.BusTerminalRepository
	notifier         RealtimeNotifier
	BusTicketSvc     BusTicketService
}

func NewNotificationService(
	platformRepo repository.PlatformRepository,
	userTerminalRepo repository.UserTerminalRepository,
	busTerminalRepo repository.BusTerminalRepository,
	notifier RealtimeNotifier,
	BusTicketSvc BusTicketService,
) *notificationService {
	return &notificationService{
		platformRepo:     platformRepo,
		userTerminalRepo: userTerminalRepo,
		busTerminalRepo:  busTerminalRepo,
		notifier:         notifier,
		BusTicketSvc:     BusTicketSvc,
	}
}

func (s *notificationService) ListAdminSelectableNotificationTypes(_ context.Context, role string) (models.AdminNotificationTypesResponse, error) {
	switch role {
	case "super_admin":
		return models.AdminNotificationTypesResponse{
			Types: []models.PassengerNotificationType{
				models.PassengerNotificationLocal,
				models.PassengerNotificationGlobal,
				models.PassengerNotificationBUSDelay,
			},
		}, nil
	case "admin":
		return models.AdminNotificationTypesResponse{
			Types: []models.PassengerNotificationType{
				models.PassengerNotificationLocal,
				models.PassengerNotificationBUSDelay,
			},
		}, nil
	default:
		return models.AdminNotificationTypesResponse{}, ErrUnsupportedNotificationRole
	}
}

func (s *notificationService) NotifyPassengers(ctx context.Context, req models.NotifyPassengersRequest) (models.NotifyPassengersResponse, error) {
	if req.LicensePatent == "" {
		return models.NotifyPassengersResponse{}, ErrLicensePatentEmpty
	}
	if req.Code == "" {
		return models.NotifyPassengersResponse{}, ErrCodeEmpty
	}

	code, err := strconv.Atoi(req.Code)
	if err != nil {
		return models.NotifyPassengersResponse{}, ErrInvalidCode
	}

	platform, err := s.platformRepo.GetByCode(ctx, code)
	if err != nil {
		return models.NotifyPassengersResponse{}, fmt.Errorf("%w: %w", ErrPlatformLookup, err)
	}
	if platform.BusTerminalID == uuid.Nil {
		return models.NotifyPassengersResponse{}, ErrPlatformMissingTerminal
	}

	platformInfo := models.PlatformInfo{
		Anden:       platform.Anden,
		Coordinates: platform.Coordinates,
	}

	payload, err := json.Marshal(platformInfo)
	if err != nil {
		return models.NotifyPassengersResponse{}, fmt.Errorf("%w: %w", ErrNotification, err)
	}

	msg := models.PassengerNotificationMessage{
		Type:    models.PassengerNotificationBUSArrival,
		Payload: payload,
	}

	groupKey := req.LicensePatent + ":" + platform.BusTerminalID.String()
	if err := s.notifier.Invoke(ctx, "SendToFrontend", groupKey, msg); err != nil {
		return models.NotifyPassengersResponse{}, fmt.Errorf("%w: %w", ErrNotification, err)
	}

	return models.NotifyPassengersResponse{
		Message: "passengers notified successfully",
	}, nil
}

func (s *notificationService) SendAdminNotification(
	ctx context.Context,
	userID uuid.UUID,
	role string,
	queryTerminalUUID string,
	req models.AdminSendNotificationRequest,
) (models.AdminSendNotificationResponse, error) {
	switch req.Type {
	case models.PassengerNotificationGlobal:
		return s.sendAdminNotificationGlobal(ctx, role, req.Payload)
	case models.PassengerNotificationLocal:
		return s.sendAdminNotificationLocal(ctx, userID, role, queryTerminalUUID, req.Payload)
	default:
		return models.AdminSendNotificationResponse{}, ErrNotificationTypeInvalid
	}
}

func (s *notificationService) sendAdminNotificationGlobal(
	ctx context.Context,
	role string,
	payloadRaw json.RawMessage,
) (models.AdminSendNotificationResponse, error) {
	if role != "super_admin" {
		return models.AdminSendNotificationResponse{}, ErrNotificationGlobalSuperAdminOnly
	}
	trimmed := bytes.TrimSpace(payloadRaw)
	if len(trimmed) == 0 {
		return models.AdminSendNotificationResponse{}, ErrNotificationPayloadEmpty
	}
	if !json.Valid(trimmed) {
		return models.AdminSendNotificationResponse{}, ErrNotificationPayloadInvalidJSON
	}

	msg := models.PassengerNotificationMessage{
		Type:    models.PassengerNotificationGlobal,
		Payload: json.RawMessage(append([]byte(nil), trimmed...)),
	}

	if err := s.notifier.Invoke(ctx, "SendToFrontendGlobal", msg); err != nil {
		return models.AdminSendNotificationResponse{}, fmt.Errorf("%w: %w", ErrNotification, err)
	}

	return models.AdminSendNotificationResponse{Message: "notification sent"}, nil
}

func (s *notificationService) sendAdminNotificationLocal(
	ctx context.Context,
	userID uuid.UUID,
	role string,
	queryTerminalUUID string,
	payloadRaw json.RawMessage,
) (models.AdminSendNotificationResponse, error) {
	if len(bytes.TrimSpace(payloadRaw)) == 0 {
		return models.AdminSendNotificationResponse{}, ErrNotificationPayloadEmpty
	}
	var localPayload models.AdminLocalNotificationPayload
	if err := json.Unmarshal(payloadRaw, &localPayload); err != nil {
		return models.AdminSendNotificationResponse{}, ErrNotificationPayloadInvalidJSON
	}
	if localPayload.Message == "" {
		return models.AdminSendNotificationResponse{}, ErrNotificationMessageEmpty
	}

	var terminalID uuid.UUID

	switch role {
	case "admin":
		uts, err := s.userTerminalRepo.GetByUserID(ctx, userID)
		if err != nil {
			return models.AdminSendNotificationResponse{}, err
		}
		switch len(uts) {
		case 0:
			return models.AdminSendNotificationResponse{}, ErrAdminNoTerminal
		case 1:
			if queryTerminalUUID != "" {
				id, perr := uuid.Parse(queryTerminalUUID)
				if perr != nil {
					return models.AdminSendNotificationResponse{}, ErrInvalidTerminalUUID
				}
				if id != uts[0].BusTerminalID {
					return models.AdminSendNotificationResponse{}, ErrTerminalNotOwned
				}
				terminalID = id
			} else {
				terminalID = uts[0].BusTerminalID
			}
		default:
			if queryTerminalUUID == "" {
				return models.AdminSendNotificationResponse{}, ErrTerminalUUIDRequiredMultiAdmin
			}
			id, perr := uuid.Parse(queryTerminalUUID)
			if perr != nil {
				return models.AdminSendNotificationResponse{}, ErrInvalidTerminalUUID
			}
			owned, exErr := s.userTerminalRepo.Exists(ctx, userID, id)
			if exErr != nil {
				return models.AdminSendNotificationResponse{}, exErr
			}
			if !owned {
				return models.AdminSendNotificationResponse{}, ErrTerminalNotOwned
			}
			if _, err := s.busTerminalRepo.GetByUUID(ctx, id); err != nil {
				if errors.Is(err, repository.ErrNotFound) {
					return models.AdminSendNotificationResponse{}, ErrTerminalNotFound
				}
				return models.AdminSendNotificationResponse{}, err
			}
			terminalID = id
		}
	case "super_admin":
		if queryTerminalUUID == "" {
			return models.AdminSendNotificationResponse{}, ErrTerminalUUIDRequired
		}
		id, err := uuid.Parse(queryTerminalUUID)
		if err != nil {
			return models.AdminSendNotificationResponse{}, ErrInvalidTerminalUUID
		}
		if _, err := s.busTerminalRepo.GetByUUID(ctx, id); err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return models.AdminSendNotificationResponse{}, ErrTerminalNotFound
			}
			return models.AdminSendNotificationResponse{}, err
		}
		terminalID = id
	default:
		return models.AdminSendNotificationResponse{}, fmt.Errorf("unsupported role for notification: %s", role)
	}

	inner, err := json.Marshal(localPayload)
	if err != nil {
		return models.AdminSendNotificationResponse{}, fmt.Errorf("%w: %w", ErrNotification, err)
	}

	msg := models.PassengerNotificationMessage{
		Type:    models.PassengerNotificationLocal,
		Payload: inner,
	}

	if err := s.notifier.Invoke(ctx, "SendToFrontend", terminalID.String(), msg); err != nil {
		return models.AdminSendNotificationResponse{}, fmt.Errorf("%w: %w", ErrNotification, err)
	}

	return models.AdminSendNotificationResponse{Message: "notification sent"}, nil
}

func (s *notificationService) NotifyBusDelay(
	ctx context.Context,
	userID uuid.UUID,
	role string,
	req models.NotifyBusDelayRequest,
) (models.NotifyBusDelayResponse, error) {
	if req.Type != models.PassengerNotificationBUSDelay {
		return models.NotifyBusDelayResponse{}, ErrBusDelayTypeInvalid
	}
	if strings.TrimSpace(req.LicensePatent) == "" {
		return models.NotifyBusDelayResponse{}, ErrBusDelayLicensePatentRequired
	}
	if strings.TrimSpace(req.StartDate) == "" {
		return models.NotifyBusDelayResponse{}, ErrBusDelayStartDateRequired
	}
	if req.Payload.TimeDelay <= 0 {
		return models.NotifyBusDelayResponse{}, ErrBusDelayTimeDelayInvalid
	}

	terminalID, err := s.resolveTerminalForBusDelay(ctx, userID, role, strings.TrimSpace(req.UUIDTerminal))
	if err != nil {
		return models.NotifyBusDelayResponse{}, err
	}

	terminal, err := s.busTerminalRepo.GetByUUID(ctx, terminalID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return models.NotifyBusDelayResponse{}, ErrTerminalNotFound
		}
		return models.NotifyBusDelayResponse{}, err
	}
	if terminal.ExternalTerminalID == nil || *terminal.ExternalTerminalID == uuid.Nil {
		return models.NotifyBusDelayResponse{}, ErrExternalTerminalNotConfigured
	}

	exists, err := s.BusTicketSvc.TripExists(ctx, *terminal.ExternalTerminalID, req.StartDate, req.LicensePatent)
	if err != nil {
		return models.NotifyBusDelayResponse{}, err
	}
	if !exists {
		return models.NotifyBusDelayResponse{}, ErrTripNotRegistered
	}

	inner, err := json.Marshal(models.NotifyBusDelayPayload{TimeDelay: req.Payload.TimeDelay})
	if err != nil {
		return models.NotifyBusDelayResponse{}, fmt.Errorf("%w: %w", ErrNotification, err)
	}

	msg := models.PassengerNotificationMessage{
		Type:    models.PassengerNotificationBUSDelay,
		Payload: inner,
	}

	groupKey := normalizeLicensePlateForDelay(req.LicensePatent) + ":" + terminalID.String()
	if err := s.notifier.Invoke(ctx, "NotifyDelayBus", groupKey, msg); err != nil {
		return models.NotifyBusDelayResponse{}, fmt.Errorf("%w: %w", ErrNotification, err)
	}

	return models.NotifyBusDelayResponse{Message: "bus delay notification sent"}, nil
}

func normalizeLicensePlateForDelay(patent string) string {
	return strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(patent), " ", ""))
}

func (s *notificationService) resolveTerminalForBusDelay(
	ctx context.Context,
	userID uuid.UUID,
	role string,
	uuidTerminal string,
) (uuid.UUID, error) {
	switch role {
	case "admin":
		uts, err := s.userTerminalRepo.GetByUserID(ctx, userID)
		if err != nil {
			return uuid.Nil, err
		}
		switch len(uts) {
		case 0:
			return uuid.Nil, ErrAdminNoTerminal
		case 1:
			if uuidTerminal != "" {
				id, perr := uuid.Parse(uuidTerminal)
				if perr != nil {
					return uuid.Nil, ErrInvalidTerminalUUID
				}
				if id != uts[0].BusTerminalID {
					return uuid.Nil, ErrTerminalNotOwned
				}
				return id, nil
			}
			return uts[0].BusTerminalID, nil
		default:
			if uuidTerminal == "" {
				return uuid.Nil, ErrBusDelayTerminalUUIDRequired
			}
			id, perr := uuid.Parse(uuidTerminal)
			if perr != nil {
				return uuid.Nil, ErrInvalidTerminalUUID
			}
			owned, exErr := s.userTerminalRepo.Exists(ctx, userID, id)
			if exErr != nil {
				return uuid.Nil, exErr
			}
			if !owned {
				return uuid.Nil, ErrTerminalNotOwned
			}
			if _, err := s.busTerminalRepo.GetByUUID(ctx, id); err != nil {
				if errors.Is(err, repository.ErrNotFound) {
					return uuid.Nil, ErrTerminalNotFound
				}
				return uuid.Nil, err
			}
			return id, nil
		}
	case "super_admin":
		if uuidTerminal == "" {
			return uuid.Nil, ErrBusDelayTerminalUUIDRequired
		}
		id, err := uuid.Parse(uuidTerminal)
		if err != nil {
			return uuid.Nil, ErrInvalidTerminalUUID
		}
		if _, err := s.busTerminalRepo.GetByUUID(ctx, id); err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return uuid.Nil, ErrTerminalNotFound
			}
			return uuid.Nil, err
		}
		return id, nil
	default:
		return uuid.Nil, fmt.Errorf("unsupported role for bus delay notification: %s", role)
	}
}
