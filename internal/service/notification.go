package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"

	"tesina/backend/internal/mail"
	"tesina/backend/internal/models"
	"tesina/backend/internal/realtime"
	"tesina/backend/internal/repository"
	"tesina/backend/internal/roles"
	"tesina/backend/internal/validators"
	errorsService "tesina/backend/internal/errors"
)

type RealtimeNotifier interface {
	Invoke(ctx context.Context, method string, args ...any) error
}

type NotificationService interface {
	NotifyPassengers(ctx context.Context, userID uuid.UUID, role string, req models.NotifyPassengersRequest) (models.NotifyPassengersResponse, error)
	SendAdminNotification(ctx context.Context, userID uuid.UUID, role string, queryTerminalUUID string, req models.AdminSendNotificationRequest) (models.AdminSendNotificationResponse, error)
	NotifyBusDelay(ctx context.Context, userID uuid.UUID, role string, req models.NotifyBusDelayRequest) (models.NotifyBusDelayResponse, error)
	ListAdminSelectableNotificationTypes(ctx context.Context, role string) (models.AdminNotificationTypesResponse, error)
	NotifyAdminCameraError(ctx context.Context, req models.CameraErrorNotifyRequest) (models.CameraErrorNotifyResponse, error)
	ListNotifications(ctx context.Context) ([]models.Notification, error)
	GetNotifications(ctx context.Context, userID uuid.UUID, role string, params models.GetNotificationsParams) (models.GetNotificationsResponse, error)
	DeleteNotification(ctx context.Context, userID uuid.UUID, role string, notificationID uuid.UUID) error
	ListAdminNotifications(ctx context.Context, userID uuid.UUID, role string, params models.ListAdminNotificationsParams) (models.ListAdminNotificationsResponse, error)
	GetAdminNotification(ctx context.Context, userID uuid.UUID, role string, notificationID uuid.UUID) (models.AdminNotificationListItem, error)
}

type notificationService struct {
	platformRepo         repository.PlatformRepository
	userTerminalRepo     repository.UserTerminalRepository
	busTerminalRepo      repository.BusTerminalRepository
	notificationRepo     repository.NotificationRepository
	awaitedTripRepo      repository.AwaitedTripRepository
	notifier             RealtimeNotifier
	hubMethods           realtime.RealtimeHubMethods
	BusTicketSvc         BusTicketService
	mailer               *mail.Mailer
	mailSiteName         string
}

func NewNotificationService(
	platformRepo repository.PlatformRepository,
	userTerminalRepo repository.UserTerminalRepository,
	busTerminalRepo repository.BusTerminalRepository,
	notificationRepo repository.NotificationRepository,
	awaitedTripRepo repository.AwaitedTripRepository,
	notifier RealtimeNotifier,
	hubMethods realtime.RealtimeHubMethods,
	BusTicketSvc BusTicketService,
	mailer *mail.Mailer,
	mailSiteName string,
) *notificationService {
	return &notificationService{
		platformRepo:         platformRepo,
		userTerminalRepo:     userTerminalRepo,
		busTerminalRepo:      busTerminalRepo,
		notificationRepo:     notificationRepo,
		awaitedTripRepo:      awaitedTripRepo,
		notifier:             notifier,
		hubMethods:           hubMethods,
		BusTicketSvc:         BusTicketSvc,
		mailer:               mailer,
		mailSiteName:         mailSiteName,
	}
}

func (s *notificationService) ListAdminSelectableNotificationTypes(_ context.Context, role string) (models.AdminNotificationTypesResponse, error) {
	switch role {
	case roles.SuperAdmin:
		return models.AdminNotificationTypesResponse{
			Types: []models.PassengerNotificationType{
				models.PassengerNotificationLocal,
				models.PassengerNotificationGlobal,
				models.PassengerNotificationBUSDelay,
				models.PassengerNotificationBUSArrival,
			},
		}, nil
	case roles.Admin:
		return models.AdminNotificationTypesResponse{
			Types: []models.PassengerNotificationType{
				models.PassengerNotificationLocal,
				models.PassengerNotificationBUSDelay,
				models.PassengerNotificationBUSArrival,
			},
		}, nil
	default:
		return models.AdminNotificationTypesResponse{}, validators.ErrUnsupportedNotificationRole
	}
}

// NotifyPassengers avisa la llegada de un colectivo. La invoca la cámara (X-API-Key, role vacío)
// o un admin/super_admin manualmente (JWT) cuando la cámara falla; el admin solo puede avisar
// en andenes de terminales que tiene asignadas.
func (s *notificationService) NotifyPassengers(ctx context.Context, userID uuid.UUID, role string, req models.NotifyPassengersRequest) (models.NotifyPassengersResponse, error) {
	code, err := validators.ValidateNotifyPassengersRequest(req)
	if err != nil {
		return models.NotifyPassengersResponse{}, err
	}

	platform, err := s.platformRepo.GetByCode(ctx, code)
	if err != nil {
		return models.NotifyPassengersResponse{}, fmt.Errorf("%w: %w", errorsService.ErrPlatformLookup, err)
	}
	if platform.BusTerminalID == uuid.Nil {
		return models.NotifyPassengersResponse{}, errorsService.ErrPlatformMissingTerminal
	}
	if role == roles.Admin {
		owned, err := s.userTerminalRepo.Exists(ctx, userID, platform.BusTerminalID)
		if err != nil {
			return models.NotifyPassengersResponse{}, err
		}
		if !owned {
			return models.NotifyPassengersResponse{}, errorsService.ErrTerminalNotOwned
		}
	}

	notifID := uuid.New()
	platformInfo := models.PlatformInfo{
		ID:          notifID,
		Anden:       platform.Anden,
		Coordinates: platform.Coordinates,
		TimeLife:    req.TimeLife,
	}

	payload, err := json.Marshal(platformInfo)
	if err != nil {
		return models.NotifyPassengersResponse{}, fmt.Errorf("%w: %w", errorsService.ErrNotification, err)
	}

	msg := models.PassengerNotificationMessage{
		Type:    models.PassengerNotificationBUSArrival,
		Payload: payload,
	}

	licensePatent := normalizeLicensePlate(req.LicensePatent)
	groupKey := licensePatent + ":" + platform.BusTerminalID.String()
	groupName := realtime.GroupPrefixFrontend + groupKey

	msgJSON, err := json.Marshal(msg)
	if err != nil {
		return models.NotifyPassengersResponse{}, fmt.Errorf("%w: %w", errorsService.ErrNotification, err)
	}
	if err := s.notificationRepo.Insert(ctx, models.Notification{
		ID:         notifID,
		GroupKey:   &groupKey,
		GroupName:  groupName,
		Expiration: time.Now().UTC().Add(time.Duration(req.TimeLife) * time.Minute),
		Date:       time.Now().UTC(),
		Payload:    msgJSON,
	}); err != nil {
		return models.NotifyPassengersResponse{}, fmt.Errorf("%w: %w", errorsService.ErrNotification, err)
	}

	if err := s.notifier.Invoke(ctx, s.hubMethods.SendToFrontend, groupName, msg); err != nil {
		return models.NotifyPassengersResponse{}, fmt.Errorf("%w: %w", errorsService.ErrNotification, err)
	}

	go s.notifyAwaitingPassengers(groupKey, licensePatent, platform.Anden, platform.Coordinates)

	return models.NotifyPassengersResponse{
		Message: "passengers notified successfully",
	}, nil
}

// notifyAwaitingPassengers marca como notificados a los pasajeros que esperaban
// el colectivo y les manda el mail de llegada (si hay SMTP configurado).
// El awaited_trip no se borra: el pasajero puede recargar el frontend y seguir
// viendo su viaje hasta que lo abandone.
func (s *notificationService) notifyAwaitingPassengers(groupKey, licensePatent, anden string, coordinates json.RawMessage) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	emails, err := s.awaitedTripRepo.MarkNotifiedByGroupKey(ctx, groupKey)
	if err != nil {
		log.Printf("bus arrival: mark awaited_trip notified for group %q: %v", groupKey, err)
		return
	}
	if len(emails) == 0 || s.mailer == nil {
		return
	}

	body, err := mail.RenderBusArrivalHTML(mail.BusArrivalEmailData{
		SiteName:      s.mailSiteName,
		LicensePatent: licensePatent,
		Anden:         anden,
		MapsURL:       mapsURLFromCoords(coordinates),
	})
	if err != nil {
		log.Printf("bus arrival email: render template: %v", err)
		return
	}

	for _, email := range emails {
		if err := s.mailer.Send(mail.SendOptions{
			To:      []string{email},
			Subject: "Tu bus ha llegado",
			Body:    body,
			IsHTML:  true,
		}); err != nil {
			log.Printf("bus arrival email: send to %s: %v", email, err)
		}
	}
}

func mapsURLFromCoords(coordinates json.RawMessage) string {
	if len(coordinates) == 0 {
		return ""
	}
	var c struct {
		Lat float64 `json:"lat"`
		Lng float64 `json:"lng"`
	}
	if err := json.Unmarshal(coordinates, &c); err != nil || (c.Lat == 0 && c.Lng == 0) {
		return ""
	}
	return fmt.Sprintf("https://www.google.com/maps?q=%.6f,%.6f", c.Lat, c.Lng)
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
		return models.AdminSendNotificationResponse{}, validators.ErrNotificationTypeInvalid
	}
}

func (s *notificationService) sendAdminNotificationGlobal(
	ctx context.Context,
	role string,
	payloadRaw json.RawMessage,
) (models.AdminSendNotificationResponse, error) {
	timeLife, err := validators.ValidateAdminGlobalNotification(role, payloadRaw)
	if err != nil {
		return models.AdminSendNotificationResponse{}, err
	}

	trimmed := bytes.TrimSpace(payloadRaw)
	notifID := uuid.New()
	merged, err := mergeJSONWithFields(trimmed, map[string]any{
		"id": notifID.String(),
	})
	if err != nil {
		return models.AdminSendNotificationResponse{}, fmt.Errorf("%w: %w", errorsService.ErrNotification, err)
	}

	msg := models.PassengerNotificationMessage{
		Type:    models.PassengerNotificationGlobal,
		Payload: merged,
	}

	groupName := realtime.GroupNameFrontendGlobal

	msgJSON, err := json.Marshal(msg)
	if err != nil {
		return models.AdminSendNotificationResponse{}, fmt.Errorf("%w: %w", errorsService.ErrNotification, err)
	}
	if err := s.notificationRepo.Insert(ctx, models.Notification{
		ID:         notifID,
		GroupKey:   nil,
		GroupName:  groupName,
		Expiration: time.Now().UTC().Add(time.Duration(timeLife) * time.Minute),
		Date:       time.Now().UTC(),
		Payload:    msgJSON,
	}); err != nil {
		return models.AdminSendNotificationResponse{}, fmt.Errorf("%w: %w", errorsService.ErrNotification, err)
	}

	if err := s.notifier.Invoke(ctx, s.hubMethods.SendToFrontendGlobal, groupName, msg); err != nil {
		return models.AdminSendNotificationResponse{}, fmt.Errorf("%w: %w", errorsService.ErrNotification, err)
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
	localPayload, err := validators.ValidateAdminLocalNotification(payloadRaw)
	if err != nil {
		return models.AdminSendNotificationResponse{}, err
	}

	notifID := uuid.New()
	localPayload.ID = notifID.String()

	var terminalID uuid.UUID

	switch role {
	case roles.Admin:
		uts, err := s.userTerminalRepo.GetByUserID(ctx, userID)
		if err != nil {
			return models.AdminSendNotificationResponse{}, err
		}
		switch len(uts) {
		case 0:
			return models.AdminSendNotificationResponse{}, errorsService.ErrAdminNoTerminal
		case 1:
			if queryTerminalUUID != "" {
				id, perr := uuid.Parse(queryTerminalUUID)
				if perr != nil {
					return models.AdminSendNotificationResponse{}, errorsService.ErrInvalidTerminalUUID
				}
				if id != uts[0].BusTerminalID {
					return models.AdminSendNotificationResponse{}, errorsService.ErrTerminalNotOwned
				}
				terminalID = id
			} else {
				terminalID = uts[0].BusTerminalID
			}
		default:
			if queryTerminalUUID == "" {
				return models.AdminSendNotificationResponse{}, errorsService.ErrTerminalUUIDRequiredMultiAdmin
			}
			id, perr := uuid.Parse(queryTerminalUUID)
			if perr != nil {
				return models.AdminSendNotificationResponse{}, errorsService.ErrInvalidTerminalUUID
			}
			owned, exErr := s.userTerminalRepo.Exists(ctx, userID, id)
			if exErr != nil {
				return models.AdminSendNotificationResponse{}, exErr
			}
			if !owned {
				return models.AdminSendNotificationResponse{}, errorsService.ErrTerminalNotOwned
			}
			if _, err := s.busTerminalRepo.GetByUUID(ctx, id); err != nil {
				if errors.Is(err, repository.ErrNotFound) {
					return models.AdminSendNotificationResponse{}, errorsService.ErrTerminalNotFound
				}
				return models.AdminSendNotificationResponse{}, err
			}
			terminalID = id
		}
	case roles.SuperAdmin:
		if queryTerminalUUID == "" {
			return models.AdminSendNotificationResponse{}, errorsService.ErrTerminalUUIDRequired
		}
		id, err := uuid.Parse(queryTerminalUUID)
		if err != nil {
			return models.AdminSendNotificationResponse{}, errorsService.ErrInvalidTerminalUUID
		}
		if _, err := s.busTerminalRepo.GetByUUID(ctx, id); err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return models.AdminSendNotificationResponse{}, errorsService.ErrTerminalNotFound
			}
			return models.AdminSendNotificationResponse{}, err
		}
		terminalID = id
	default:
		return models.AdminSendNotificationResponse{}, fmt.Errorf("unsupported role for notification: %s", role)
	}

	inner, err := json.Marshal(localPayload)
	if err != nil {
		return models.AdminSendNotificationResponse{}, fmt.Errorf("%w: %w", errorsService.ErrNotification, err)
	}

	msg := models.PassengerNotificationMessage{
		Type:    models.PassengerNotificationLocal,
		Payload: inner,
	}

	groupKey := terminalID.String()
	groupName := realtime.GroupPrefixFrontend + groupKey

	msgJSON, err := json.Marshal(msg)
	if err != nil {
		return models.AdminSendNotificationResponse{}, fmt.Errorf("%w: %w", errorsService.ErrNotification, err)
	}
	if err := s.notificationRepo.Insert(ctx, models.Notification{
		ID:         notifID,
		GroupKey:   &groupKey,
		GroupName:  groupName,
		Expiration: time.Now().UTC().Add(time.Duration(localPayload.TimeLife) * time.Minute),
		Date:       time.Now().UTC(),
		Payload:    msgJSON,
	}); err != nil {
		return models.AdminSendNotificationResponse{}, fmt.Errorf("%w: %w", errorsService.ErrNotification, err)
	}

	if err := s.notifier.Invoke(ctx, s.hubMethods.SendToFrontend, groupName, msg); err != nil {
		return models.AdminSendNotificationResponse{}, fmt.Errorf("%w: %w", errorsService.ErrNotification, err)
	}

	return models.AdminSendNotificationResponse{Message: "notification sent"}, nil
}

func (s *notificationService) NotifyBusDelay(
	ctx context.Context,
	userID uuid.UUID,
	role string,
	req models.NotifyBusDelayRequest,
) (models.NotifyBusDelayResponse, error) {
	if err := validators.ValidateNotifyBusDelayRequest(req); err != nil {
		return models.NotifyBusDelayResponse{}, err
	}

	terminalID, err := s.resolveTerminalForBusDelay(ctx, userID, role, strings.TrimSpace(req.UUIDTerminal))
	if err != nil {
		return models.NotifyBusDelayResponse{}, err
	}

	terminal, err := s.busTerminalRepo.GetByUUID(ctx, terminalID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return models.NotifyBusDelayResponse{}, errorsService.ErrTerminalNotFound
		}
		return models.NotifyBusDelayResponse{}, err
	}
	if terminal.ExternalTerminalID == nil || *terminal.ExternalTerminalID == uuid.Nil {
		return models.NotifyBusDelayResponse{}, errorsService.ErrExternalTerminalNotConfigured
	}

	exists, err := s.BusTicketSvc.TripExists(ctx, *terminal.ExternalTerminalID, req.StartDate, req.LicensePatent)

	if err != nil {
		return models.NotifyBusDelayResponse{}, err
	}
	if !exists {
		return models.NotifyBusDelayResponse{}, errorsService.ErrTripNotRegistered
	}

	notifID := uuid.New()
	inner, err := json.Marshal(models.NotifyBusDelayPayload{
		ID:            notifID.String(),
		LicensePatent: req.LicensePatent,
		TimeDelay:     req.Payload.TimeDelay,
		TimeLife:      req.Payload.TimeLife,
	})
	if err != nil {
		return models.NotifyBusDelayResponse{}, fmt.Errorf("%w: %w", errorsService.ErrNotification, err)
	}

	msg := models.PassengerNotificationMessage{
		Type:    models.PassengerNotificationBUSDelay,
		Payload: inner,
	}

	normalizedPatent := normalizeLicensePlateForDelay(req.LicensePatent)

	compositeKey := normalizedPatent + ":" + terminalID.String()
	groupName := realtime.GroupPrefixFrontend + compositeKey

	msgJSON, err := json.Marshal(msg)
	if err != nil {
		return models.NotifyBusDelayResponse{}, fmt.Errorf("%w: %w", errorsService.ErrNotification, err)
	}
	if err := s.notificationRepo.Insert(ctx, models.Notification{
		ID:         notifID,
		GroupKey:   &compositeKey,
		GroupName:  groupName,
		Expiration: time.Now().UTC().Add(time.Duration(req.Payload.TimeLife) * time.Minute),
		Date:       time.Now().UTC(),
		Payload:    msgJSON,
	}); err != nil {
		return models.NotifyBusDelayResponse{}, fmt.Errorf("%w: %w", errorsService.ErrNotification, err)
	}

	if err := s.notifier.Invoke(ctx, s.hubMethods.NotifyDelayBus, groupName, msg); err != nil {
		return models.NotifyBusDelayResponse{}, fmt.Errorf("%w: %w", errorsService.ErrNotification, err)
	}

	return models.NotifyBusDelayResponse{Message: "bus delay notification sent"}, nil
}

func normalizeLicensePlateForDelay(patent string) string {
	return strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(patent), " ", ""))
}

func (s *notificationService) NotifyAdminCameraError(
	ctx context.Context,
	req models.CameraErrorNotifyRequest,
) (models.CameraErrorNotifyResponse, error) {
	code, err := validators.ValidateCameraErrorRequest(req)
	if err != nil {
		return models.CameraErrorNotifyResponse{}, err
	}

	platform, err := s.platformRepo.GetByCode(ctx, code)
	if err != nil {
		return models.CameraErrorNotifyResponse{}, fmt.Errorf("%w: %w", errorsService.ErrPlatformLookup, err)
	}
	if platform.BusTerminalID == uuid.Nil {
		return models.CameraErrorNotifyResponse{}, errorsService.ErrPlatformMissingTerminal
	}

	notifID := uuid.New()
	cameraPayload := models.CameraErrorNotifyPayload{
		ID:       notifID.String(),
		Message:  strings.TrimSpace(req.Payload.Message),
		TimeLife: req.Payload.TimeLife,
	}
	inner, err := json.Marshal(cameraPayload)
	if err != nil {
		return models.CameraErrorNotifyResponse{}, fmt.Errorf("%w: %w", errorsService.ErrNotification, err)
	}

	msg := models.PassengerNotificationMessage{
		Type:    models.PassengerNotificationCAMERA,
		Payload: inner,
	}

	groupKey := platform.BusTerminalID.String()
	groupName := realtime.GroupPrefixFrontendAdmin + groupKey

	msgJSON, err := json.Marshal(msg)
	if err != nil {
		return models.CameraErrorNotifyResponse{}, fmt.Errorf("%w: %w", errorsService.ErrNotification, err)
	}
	if err := s.notificationRepo.Insert(ctx, models.Notification{
		ID:         notifID,
		GroupKey:   &groupKey,
		GroupName:  groupName,
		Expiration: time.Now().UTC().Add(time.Duration(req.Payload.TimeLife) * time.Minute),
		Date:       time.Now().UTC(),
		Payload:    msgJSON,
	}); err != nil {
		return models.CameraErrorNotifyResponse{}, fmt.Errorf("%w: %w", errorsService.ErrNotification, err)
	}

	if err := s.notifier.Invoke(ctx, s.hubMethods.NotifyAdminFromCamera, groupName, msg); err != nil {
		return models.CameraErrorNotifyResponse{}, fmt.Errorf("%w: %w", errorsService.ErrNotification, err)
	}

	return models.CameraErrorNotifyResponse{
		Type:    models.PassengerNotificationCAMERA,
		Payload: cameraPayload,
	}, nil
}

func (s *notificationService) ListNotifications(ctx context.Context) ([]models.Notification, error) {
	return s.notificationRepo.List(ctx)
}

func (s *notificationService) DeleteNotification(ctx context.Context, userID uuid.UUID, role string, notificationID uuid.UUID) error {
	switch role {
	case roles.User:
		return errorsService.ErrUserCannotDeleteNotification

	case roles.SuperAdmin:
		n, err := s.notificationRepo.GetByID(ctx, notificationID)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return errorsService.ErrNotificationNotFound
			}
			return err
		}
		if err := s.notifier.Invoke(ctx, s.hubMethods.DeleteNotification, notificationID.String(), n.GroupName); err != nil {
			return fmt.Errorf("%w: %w", errorsService.ErrNotification, err)
		}
		return s.notificationRepo.Delete(ctx, notificationID)

	case roles.Admin:
		n, err := s.notificationRepo.GetByID(ctx, notificationID)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return errorsService.ErrNotificationNotFound
			}
			return err
		}
		// Global notifications (null group_key) are super_admin-only
		if n.GroupKey == nil {
			return errorsService.ErrNotificationDeleteForbidden
		}
		uts, err := s.userTerminalRepo.GetByUserID(ctx, userID)
		if err != nil {
			return err
		}
		if len(uts) == 0 {
			return errorsService.ErrAdminNoTerminal
		}
		for _, ut := range uts {
			if strings.Contains(*n.GroupKey, ut.BusTerminalID.String()) {
				if err := s.notifier.Invoke(ctx, s.hubMethods.DeleteNotification, notificationID.String(), n.GroupName); err != nil {
					return fmt.Errorf("%w: %w", errorsService.ErrNotification, err)
				}
				return s.notificationRepo.Delete(ctx, notificationID)
			}
		}
		return errorsService.ErrNotificationDeleteForbidden

	default:
		return errorsService.ErrNotificationDeleteForbidden
	}
}

const (
	defaultAdminNotificationsLimit = 10
	maxAdminNotificationsLimit     = 100
)

// ListAdminNotifications lista las notificaciones según el rol:
// super_admin ve todas (incluidas las globales); admin solo las de sus terminales.
func (s *notificationService) ListAdminNotifications(
	ctx context.Context,
	userID uuid.UUID,
	role string,
	params models.ListAdminNotificationsParams,
) (models.ListAdminNotificationsResponse, error) {
	page, limit, order := normalizePagination(params.Page, params.Limit, params.Order, defaultAdminNotificationsLimit, maxAdminNotificationsLimit)

	var f models.AdminNotificationFilters

	if params.Type != "" {
		t := models.PassengerNotificationType(strings.ToUpper(strings.TrimSpace(params.Type)))
		if !isKnownNotificationType(t) {
			return models.ListAdminNotificationsResponse{}, errorsService.ErrNotificationListTypeInvalid
		}
		f.Type = &t
	}

	switch status := strings.ToLower(strings.TrimSpace(params.Status)); status {
	case "", "all":
	case "active", "expired":
		f.Status = status
	default:
		return models.ListAdminNotificationsResponse{}, errorsService.ErrNotificationStatusInvalid
	}

	var terminalFilter *uuid.UUID
	if raw := strings.TrimSpace(params.TerminalUUID); raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil {
			return models.ListAdminNotificationsResponse{}, errorsService.ErrInvalidTerminalUUIDFilter
		}
		terminalFilter = &id
	}

	switch role {
	case roles.SuperAdmin:
		if terminalFilter != nil {
			f.TerminalIDs = []string{terminalFilter.String()}
		}

	case roles.Admin:
		uts, err := s.userTerminalRepo.GetByUserID(ctx, userID)
		if err != nil {
			return models.ListAdminNotificationsResponse{}, fmt.Errorf("failed to list admin terminals: %w", err)
		}
		if len(uts) == 0 {
			return models.ListAdminNotificationsResponse{}, errorsService.ErrAdminNoTerminal
		}
		if terminalFilter != nil {
			owned := false
			for _, ut := range uts {
				if ut.BusTerminalID == *terminalFilter {
					owned = true
					break
				}
			}
			if !owned {
				return models.ListAdminNotificationsResponse{}, errorsService.ErrTerminalNotOwned
			}
			f.TerminalIDs = []string{terminalFilter.String()}
		} else {
			f.TerminalIDs = make([]string, len(uts))
			for i, ut := range uts {
				f.TerminalIDs[i] = ut.BusTerminalID.String()
			}
		}

	default:
		return models.ListAdminNotificationsResponse{}, errorsService.ErrNotificationListForbidden
	}

	total, err := s.notificationRepo.CountAdmin(ctx, f)
	if err != nil {
		return models.ListAdminNotificationsResponse{}, fmt.Errorf("failed to count notifications: %w", err)
	}

	rows, err := s.notificationRepo.ListAdminPaginated(ctx, f, limit, (page-1)*limit, order)
	if err != nil {
		return models.ListAdminNotificationsResponse{}, fmt.Errorf("failed to list notifications: %w", err)
	}

	items, err := s.toAdminNotificationItems(ctx, rows)
	if err != nil {
		return models.ListAdminNotificationsResponse{}, err
	}

	next, prev := pageLinks(total, page, limit)

	return models.ListAdminNotificationsResponse{
		Notifications: items,
		Page:          page,
		Next:          next,
		Prev:          prev,
		Elements:      len(items),
		TotalElements: total,
	}, nil
}

// GetAdminNotification devuelve una notificación si el rol puede verla:
// super_admin cualquiera; admin solo las de sus terminales.
func (s *notificationService) GetAdminNotification(
	ctx context.Context,
	userID uuid.UUID,
	role string,
	notificationID uuid.UUID,
) (models.AdminNotificationListItem, error) {
	if role != roles.SuperAdmin && role != roles.Admin {
		return models.AdminNotificationListItem{}, errorsService.ErrNotificationListForbidden
	}

	n, err := s.notificationRepo.GetByID(ctx, notificationID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return models.AdminNotificationListItem{}, errorsService.ErrNotificationNotFound
		}
		return models.AdminNotificationListItem{}, fmt.Errorf("failed to get notification: %w", err)
	}

	if role == roles.Admin {
		terminalID := terminalIDFromGroupKey(n.GroupKey)
		if terminalID == nil {
			return models.AdminNotificationListItem{}, errorsService.ErrNotificationNotFound
		}
		owned, err := s.userTerminalRepo.Exists(ctx, userID, *terminalID)
		if err != nil {
			return models.AdminNotificationListItem{}, fmt.Errorf("failed to check admin terminal: %w", err)
		}
		if !owned {
			return models.AdminNotificationListItem{}, errorsService.ErrNotificationNotFound
		}
	}

	items, err := s.toAdminNotificationItems(ctx, []models.Notification{n})
	if err != nil {
		return models.AdminNotificationListItem{}, err
	}
	return items[0], nil
}

// toAdminNotificationItems arma los ítems del listado de admin, resolviendo
// el nombre de cada terminal con una sola consulta.
func (s *notificationService) toAdminNotificationItems(ctx context.Context, rows []models.Notification) ([]models.AdminNotificationListItem, error) {
	terminalIDs := make([]*uuid.UUID, len(rows))
	var lookup []uuid.UUID
	seen := make(map[uuid.UUID]bool)
	for i, n := range rows {
		terminalIDs[i] = terminalIDFromGroupKey(n.GroupKey)
		if id := terminalIDs[i]; id != nil && !seen[*id] {
			seen[*id] = true
			lookup = append(lookup, *id)
		}
	}
	names := make(map[uuid.UUID]string, len(lookup))
	if len(lookup) > 0 {
		terminals, err := s.busTerminalRepo.ListByUUIDs(ctx, lookup)
		if err != nil {
			return nil, fmt.Errorf("failed to list notification terminals: %w", err)
		}
		for _, t := range terminals {
			names[t.UUID] = t.Name
		}
	}

	now := time.Now().UTC()
	items := make([]models.AdminNotificationListItem, len(rows))
	for i, n := range rows {
		var stored models.PassengerNotificationMessage
		if err := json.Unmarshal(n.Payload, &stored); err != nil {
			stored.Payload = n.Payload
		}
		var terminal *models.ProfileTerminalRef
		if id := terminalIDs[i]; id != nil {
			terminal = &models.ProfileTerminalRef{UUID: *id, Name: names[*id]}
		}
		items[i] = models.AdminNotificationListItem{
			ID:         n.ID,
			Type:       stored.Type,
			Terminal:   terminal,
			Date:       n.Date,
			Expiration: n.Expiration,
			Expired:    !n.Expiration.After(now),
			Payload:    stored.Payload,
		}
	}
	return items, nil
}

func isKnownNotificationType(t models.PassengerNotificationType) bool {
	switch t {
	case models.PassengerNotificationBUSArrival,
		models.PassengerNotificationBUSDelay,
		models.PassengerNotificationLocal,
		models.PassengerNotificationGlobal,
		models.PassengerNotificationCAMERA:
		return true
	}
	return false
}

// terminalIDFromGroupKey extrae la terminal de un group_key ("<tid>" o "<patente>:<tid>").
// Devuelve nil para notificaciones globales o claves inválidas.
func terminalIDFromGroupKey(groupKey *string) *uuid.UUID {
	if groupKey == nil {
		return nil
	}
	key := *groupKey
	if i := strings.LastIndex(key, ":"); i >= 0 {
		key = key[i+1:]
	}
	id, err := uuid.Parse(key)
	if err != nil {
		return nil
	}
	return &id
}

func mergeJSONWithFields(base json.RawMessage, extra map[string]any) (json.RawMessage, error) {
	var m map[string]any
	if err := json.Unmarshal(base, &m); err != nil {
		return nil, err
	}
	for k, v := range extra {
		m[k] = v
	}
	return json.Marshal(m)
}

func (s *notificationService) GetNotifications(
	ctx context.Context,
	userID uuid.UUID,
	role string,
	params models.GetNotificationsParams,
) (models.GetNotificationsResponse, error) {
	var f models.NotificationFilters

	switch role {
	case roles.SuperAdmin:
		if params.LicensePlate != "" {
			plate := normalizeLicensePlateForDelay(params.LicensePlate)
			f.GroupKeyLike = []string{plate + ":%"}
		}

	case roles.Admin:
		uts, err := s.userTerminalRepo.GetByUserID(ctx, userID)
		if err != nil {
			return models.GetNotificationsResponse{}, err
		}
		if len(uts) == 0 {
			return models.GetNotificationsResponse{}, errorsService.ErrAdminNoTerminal
		}
		tids := make([]string, len(uts))
		for i, ut := range uts {
			tids[i] = ut.BusTerminalID.String()
		}
		if params.LicensePlate != "" {
			plate := normalizeLicensePlateForDelay(params.LicensePlate)
			composites := make([]string, len(tids))
			for i, tid := range tids {
				composites[i] = plate + ":" + tid
			}
			f.GroupKeyExact = append(tids, composites...)
		} else {
			f.GroupKeyExact = tids
			likes := make([]string, len(tids))
			for i, tid := range tids {
				likes[i] = "%:" + tid
			}
			f.GroupKeyLike = likes
		}

	default: // user / passenger
		tid, err := validators.ValidateGetNotificationsUserRole(params)
		if err != nil {
			return models.GetNotificationsResponse{}, err
		}
		tidStr := tid.String()
		f.GroupKeyIsNull = true
		f.GroupKeyExact = []string{tidStr}
		if params.LicensePlate != "" {
			plate := normalizeLicensePlateForDelay(params.LicensePlate)
			f.GroupKeyExact = []string{tidStr, plate + ":" + tidStr}
		} else {
			f.GroupKeyLike = []string{"%:" + tidStr}
		}
		f.ExcludeAdminGroups = true
	}

	if err := applyCommonFilters(params, &f); err != nil {
		return models.GetNotificationsResponse{}, err
	}

	rows, total, err := s.notificationRepo.ListWithFilters(ctx, f)
	if err != nil {
		return models.GetNotificationsResponse{}, err
	}

	items := make([]models.NotificationResponseItem, len(rows))
	for i, n := range rows {
		items[i] = models.NotificationResponseItem{
			ID:         n.ID,
			Expiration: n.Expiration.Format("2006-01-02 15:04:05"),
			Date:       n.Date.Format("2006-01-02"),
			Data:       stripPayloadID(n.Payload),
		}
	}

	limit := f.Limit
	if limit <= 0 {
		limit = 10
	}
	totalPages := 0
	if total > 0 {
		totalPages = int((total + int64(limit) - 1) / int64(limit))
	}

	return models.GetNotificationsResponse{
		TotalPages:    totalPages,
		NumberPage:    f.Offset/limit + 1,
		Notifications: items,
	}, nil
}

func applyCommonFilters(params models.GetNotificationsParams, f *models.NotificationFilters) error {
	if params.NotificationType != "" {
		t := models.PassengerNotificationType(params.NotificationType)
		if !isKnownNotificationType(t) {
			return validators.ErrNotificationTypeInvalid
		}
		f.NotificationType = &t
	}
	if params.ExpirationFilter == "true" {
		v := true
		f.OnlyExpired = &v
	}
	if params.StartDate != "" {
		t, err := time.Parse("2006-01-02", params.StartDate)
		if err != nil {
			return validators.ErrInvalidStartDate
		}
		f.StartDate = &t
		if params.EndDate != "" {
			t2, err := time.Parse("2006-01-02", params.EndDate)
			if err != nil {
				return validators.ErrInvalidEndDate
			}
			if t2.Before(t) {
				return validators.ErrEndDateBeforeStart
			}
			f.EndDate = &t2
		}
	}
	f.Limit = params.Limit
	f.Offset = params.Offset
	return nil
}

func stripPayloadID(raw json.RawMessage) json.RawMessage {
	var outer struct {
		Type    string          `json:"type"`
		Payload json.RawMessage `json:"payload"`
	}
	if err := json.Unmarshal(raw, &outer); err != nil {
		return raw
	}
	var inner map[string]any
	if err := json.Unmarshal(outer.Payload, &inner); err != nil {
		return raw
	}
	delete(inner, "id")
	innerClean, err := json.Marshal(inner)
	if err != nil {
		return raw
	}
	result, err := json.Marshal(map[string]any{
		"type":    outer.Type,
		"payload": json.RawMessage(innerClean),
	})
	if err != nil {
		return raw
	}
	return result
}

func (s *notificationService) resolveTerminalForBusDelay(
	ctx context.Context,
	userID uuid.UUID,
	role string,
	uuidTerminal string,
) (uuid.UUID, error) {
	switch role {
	case roles.Admin:
		uts, err := s.userTerminalRepo.GetByUserID(ctx, userID)
		if err != nil {
			return uuid.Nil, err
		}
		switch len(uts) {
		case 0:
			return uuid.Nil, errorsService.ErrAdminNoTerminal
		case 1:
			if uuidTerminal != "" {
				id, perr := uuid.Parse(uuidTerminal)
				if perr != nil {
					return uuid.Nil, errorsService.ErrInvalidTerminalUUID
				}
				if id != uts[0].BusTerminalID {
					return uuid.Nil, errorsService.ErrTerminalNotOwned
				}
				return id, nil
			}
			return uts[0].BusTerminalID, nil
		default:
			if uuidTerminal == "" {
				return uuid.Nil, errorsService.ErrBusDelayTerminalUUIDRequired
			}
			id, perr := uuid.Parse(uuidTerminal)
			if perr != nil {
				return uuid.Nil, errorsService.ErrInvalidTerminalUUID
			}
			owned, exErr := s.userTerminalRepo.Exists(ctx, userID, id)
			if exErr != nil {
				return uuid.Nil, exErr
			}
			if !owned {
				return uuid.Nil, errorsService.ErrTerminalNotOwned
			}
			if _, err := s.busTerminalRepo.GetByUUID(ctx, id); err != nil {
				if errors.Is(err, repository.ErrNotFound) {
					return uuid.Nil, errorsService.ErrTerminalNotFound
				}
				return uuid.Nil, err
			}
			return id, nil
		}
	case roles.SuperAdmin:
		if uuidTerminal == "" {
			return uuid.Nil, errorsService.ErrBusDelayTerminalUUIDRequired
		}
		id, err := uuid.Parse(uuidTerminal)
		if err != nil {
			return uuid.Nil, errorsService.ErrInvalidTerminalUUID
		}
		if _, err := s.busTerminalRepo.GetByUUID(ctx, id); err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return uuid.Nil, errorsService.ErrTerminalNotFound
			}
			return uuid.Nil, err
		}
		return id, nil
	default:
		return uuid.Nil, fmt.Errorf("unsupported role for bus delay notification: %s", role)
	}
}
