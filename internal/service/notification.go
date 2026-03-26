package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/google/uuid"

	"tesina/backend/internal/models"
	"tesina/backend/internal/repository"
)

type RealtimeNotifier interface {
	Invoke(ctx context.Context, method string, args ...any) error
}

type NotificationService interface {
	NotifyPassengers(ctx context.Context, req models.NotifyPassengersRequest) (models.NotifyPassengersResponse, error)
}

type notificationService struct {
	platformRepo repository.PlatformRepository
	notifier     RealtimeNotifier
}

func NewNotificationService(platformRepo repository.PlatformRepository, notifier RealtimeNotifier) *notificationService {
	return &notificationService{
		platformRepo: platformRepo,
		notifier:     notifier,
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
