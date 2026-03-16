package service

import (
	"context"
	"fmt"
	"strconv"

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

	platformInfo := models.PlatformInfo{
		Anden:       platform.Anden,
		Coordinates: platform.Coordinates,
	}

	if err := s.notifier.Invoke(ctx, "SendToFrontend", req.LicensePatent, platformInfo); err != nil {
		return models.NotifyPassengersResponse{}, fmt.Errorf("%w: %w", ErrNotification, err)
	}

	return models.NotifyPassengersResponse{
		Message: "passengers notified successfully",
	}, nil
}
