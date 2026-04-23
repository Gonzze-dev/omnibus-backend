package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

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
	NotifyAdminCameraError(ctx context.Context, req models.CameraErrorNotifyRequest) (models.CameraErrorNotifyResponse, error)
	ListNotifications(ctx context.Context) ([]models.Notification, error)
	GetNotifications(ctx context.Context, userID uuid.UUID, role string, params models.GetNotificationsParams) (models.GetNotificationsResponse, error)
}

type notificationService struct {
	platformRepo         repository.PlatformRepository
	userTerminalRepo     repository.UserTerminalRepository
	busTerminalRepo      repository.BusTerminalRepository
	notificationRepo     repository.NotificationRepository
	notifier             RealtimeNotifier
	hubMethods           RealtimeHubMethods
	BusTicketSvc         BusTicketService
}

func NewNotificationService(
	platformRepo repository.PlatformRepository,
	userTerminalRepo repository.UserTerminalRepository,
	busTerminalRepo repository.BusTerminalRepository,
	notificationRepo repository.NotificationRepository,
	notifier RealtimeNotifier,
	hubMethods RealtimeHubMethods,
	BusTicketSvc BusTicketService,
) *notificationService {
	return &notificationService{
		platformRepo:         platformRepo,
		userTerminalRepo:     userTerminalRepo,
		busTerminalRepo:      busTerminalRepo,
		notificationRepo:     notificationRepo,
		notifier:             notifier,
		hubMethods:           hubMethods,
		BusTicketSvc:         BusTicketSvc,
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

	if req.TimeLife <= 0 {
		return models.NotifyPassengersResponse{}, ErrNotificationTimeLifeInvalid
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
		ID:          uuid.New(),
		Anden:       platform.Anden,
		Coordinates: platform.Coordinates,
		TimeLife:    req.TimeLife,
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
	if err := s.notifier.Invoke(ctx, s.hubMethods.SendToFrontend, groupKey, msg); err != nil {
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

	var tmp struct {
		TimeLife int `json:"time_life"`
	}
	if err := json.Unmarshal(trimmed, &tmp); err != nil || tmp.TimeLife <= 0 {
		return models.AdminSendNotificationResponse{}, ErrNotificationTimeLifeInvalid
	}

	merged, err := mergeJSONWithFields(trimmed, map[string]any{
		"id": uuid.New().String(),
	})
	if err != nil {
		return models.AdminSendNotificationResponse{}, fmt.Errorf("%w: %w", ErrNotification, err)
	}

	msg := models.PassengerNotificationMessage{
		Type:    models.PassengerNotificationGlobal,
		Payload: merged,
	}

	if err := s.notifier.Invoke(ctx, s.hubMethods.SendToFrontendGlobal, msg); err != nil {
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
	if localPayload.TimeLife <= 0 {
		return models.AdminSendNotificationResponse{}, ErrNotificationTimeLifeInvalid
	}
	localPayload.ID = uuid.New().String()

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

	if err := s.notifier.Invoke(ctx, s.hubMethods.SendToFrontend, terminalID.String(), msg); err != nil {
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
	if req.Payload.TimeLife <= 0 {
		return models.NotifyBusDelayResponse{}, ErrNotificationTimeLifeInvalid
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

	inner, err := json.Marshal(models.NotifyBusDelayPayload{
		ID:        uuid.New().String(),
		TimeDelay: req.Payload.TimeDelay,
		TimeLife:  req.Payload.TimeLife,
	})
	if err != nil {
		return models.NotifyBusDelayResponse{}, fmt.Errorf("%w: %w", ErrNotification, err)
	}

	msg := models.PassengerNotificationMessage{
		Type:    models.PassengerNotificationBUSDelay,
		Payload: inner,
	}

	normalizedPatent := normalizeLicensePlateForDelay(req.LicensePatent)

	compositeKey :=  normalizedPatent + ":" + terminalID.String()

	keys := models.NotifyDelayBusKeys{
		Key:           compositeKey,
		LicensePatent: normalizedPatent,
		TerminalID:    terminalID.String(),
	}
	if err := s.notifier.Invoke(ctx, s.hubMethods.NotifyDelayBus, keys, msg); err != nil {
		return models.NotifyBusDelayResponse{}, fmt.Errorf("%w: %w", ErrNotification, err)
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
	if req.Type != models.PassengerNotificationCAMERA {
		return models.CameraErrorNotifyResponse{}, ErrCameraNotificationTypeInvalid
	}
	if strings.TrimSpace(req.CodeCamera) == "" {
		return models.CameraErrorNotifyResponse{}, ErrCodeCameraEmpty
	}
	code, err := strconv.Atoi(strings.TrimSpace(req.CodeCamera))
	if err != nil {
		return models.CameraErrorNotifyResponse{}, ErrCodeCameraInvalid
	}
	if strings.TrimSpace(req.Payload.Message) == "" {
		return models.CameraErrorNotifyResponse{}, ErrCameraErrorMessageEmpty
	}
	if req.Payload.TimeLife <= 0 {
		return models.CameraErrorNotifyResponse{}, ErrNotificationTimeLifeInvalid
	}

	platform, err := s.platformRepo.GetByCode(ctx, code)
	if err != nil {
		return models.CameraErrorNotifyResponse{}, fmt.Errorf("%w: %w", ErrPlatformLookup, err)
	}
	if platform.BusTerminalID == uuid.Nil {
		return models.CameraErrorNotifyResponse{}, ErrPlatformMissingTerminal
	}

	cameraPayload := models.CameraErrorNotifyPayload{
		ID:       uuid.New().String(),
		Message:  strings.TrimSpace(req.Payload.Message),
		TimeLife: req.Payload.TimeLife,
	}
	inner, err := json.Marshal(cameraPayload)
	if err != nil {
		return models.CameraErrorNotifyResponse{}, fmt.Errorf("%w: %w", ErrNotification, err)
	}

	msg := models.PassengerNotificationMessage{
		Type:    models.PassengerNotificationCAMERA,
		Payload: inner,
	}

	if err := s.notifier.Invoke(ctx, s.hubMethods.NotifyAdminFromCamera, platform.BusTerminalID.String(), msg); err != nil {
		return models.CameraErrorNotifyResponse{}, fmt.Errorf("%w: %w", ErrNotification, err)
	}

	return models.CameraErrorNotifyResponse{
		Type:    models.PassengerNotificationCAMERA,
		Payload: cameraPayload,
	}, nil
}

func (s *notificationService) ListNotifications(ctx context.Context) ([]models.Notification, error) {
	return s.notificationRepo.List(ctx)
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
	case "super_admin":
		if params.LicensePlate != "" {
			plate := normalizeLicensePlateForDelay(params.LicensePlate)
			f.GroupKeyLike = []string{plate + ":%"}
		}

	case "admin":
		uts, err := s.userTerminalRepo.GetByUserID(ctx, userID)
		if err != nil {
			return models.GetNotificationsResponse{}, err
		}
		if len(uts) == 0 {
			return models.GetNotificationsResponse{}, ErrAdminNoTerminal
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
		if params.TerminalID == "" {
			return models.GetNotificationsResponse{}, ErrTerminalIDRequired
		}
		tid, err := uuid.Parse(params.TerminalID)
		if err != nil {
			return models.GetNotificationsResponse{}, ErrInvalidTerminalID
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
		switch t {
		case models.PassengerNotificationBUSArrival,
			models.PassengerNotificationBUSDelay,
			models.PassengerNotificationLocal,
			models.PassengerNotificationGlobal,
			models.PassengerNotificationCAMERA:
			f.NotificationType = &t
		default:
			return ErrNotificationTypeInvalid
		}
	}
	if params.ExpirationFilter == "true" {
		v := true
		f.OnlyExpired = &v
	}
	if params.StartDate != "" {
		t, err := time.Parse("2006-01-02", params.StartDate)
		if err != nil {
			return ErrInvalidStartDate
		}
		f.StartDate = &t
		if params.EndDate != "" {
			t2, err := time.Parse("2006-01-02", params.EndDate)
			if err != nil {
				return ErrInvalidEndDate
			}
			if t2.Before(t) {
				return ErrEndDateBeforeStart
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
