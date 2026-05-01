package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"tesina/backend/internal/models"
	"tesina/backend/internal/repository"
	"tesina/backend/internal/service"
	"tesina/backend/internal/validators"
	errorsService "tesina/backend/internal/errors"
)

type NotificationHandler struct {
	svc service.NotificationService
}

func NewNotificationHandler(svc service.NotificationService) *NotificationHandler {
	return &NotificationHandler{svc: svc}
}

func (h *NotificationHandler) NotifyPassengers(c echo.Context) error {
	var req models.NotifyPassengersRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	resp, err := h.svc.NotifyPassengers(c.Request().Context(), req)
	if err != nil {
		return mapNotificationError(err)
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *NotificationHandler) ListAdminNotificationTypes(c echo.Context) error {
	role, _ := c.Get("role").(string)
	resp, err := h.svc.ListAdminSelectableNotificationTypes(c.Request().Context(), role)
	if err != nil {
		if errors.Is(err, validators.ErrUnsupportedNotificationRole) {
			return echo.NewHTTPError(http.StatusForbidden, err.Error())
		}
		return err
	}
	return c.JSON(http.StatusOK, resp)
}

func (h *NotificationHandler) SendAdminNotification(c echo.Context) error {

	userID, ok := c.Get("user_id").(uuid.UUID)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid user context")
	}
	role, _ := c.Get("role").(string)
	queryTerminalUUID := c.QueryParam("terminaluuid")

	var req models.AdminSendNotificationRequest

	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	resp, err := h.svc.SendAdminNotification(c.Request().Context(), userID, role, queryTerminalUUID, req)
	if err != nil {
		return mapAdminLocalNotificationError(err)
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *NotificationHandler) NotifyCameraError(c echo.Context) error {
	var req models.CameraErrorNotifyRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	resp, err := h.svc.NotifyAdminCameraError(c.Request().Context(), req)
	if err != nil {
		return mapCameraErrorNotifyError(err)
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *NotificationHandler) GetNotifications(c echo.Context) error {
	userID, ok := c.Get("user_id").(uuid.UUID)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid user context")
	}
	role, _ := c.Get("role").(string)

	limit := 10
	offset := 0
	if raw := c.QueryParam("limit"); raw != "" {
		if v, err := strconv.Atoi(raw); err == nil && v > 0 {
			limit = v
		}
	}
	if raw := c.QueryParam("offset"); raw != "" {
		if v, err := strconv.Atoi(raw); err == nil && v >= 0 {
			offset = v
		}
	}

	params := models.GetNotificationsParams{
		TerminalID:       c.QueryParam("terminalID"),
		NotificationType: c.QueryParam("notification_type"),
		ExpirationFilter: c.QueryParam("expirated_notifications"),
		LicensePlate:     c.QueryParam("license_plate"),
		StartDate:        c.QueryParam("start_date"),
		EndDate:          c.QueryParam("end_date"),
		Limit:            limit,
		Offset:           offset,
	}

	resp, err := h.svc.GetNotifications(c.Request().Context(), userID, role, params)
	if err != nil {
		return mapGetNotificationsError(err)
	}
	return c.JSON(http.StatusOK, resp)
}

func mapGetNotificationsError(err error) error {
	switch {
	case errors.Is(err, validators.ErrTerminalIDRequired),
		errors.Is(err, validators.ErrInvalidTerminalID),
		errors.Is(err, validators.ErrNotificationTypeInvalid),
		errors.Is(err, validators.ErrInvalidStartDate),
		errors.Is(err, validators.ErrInvalidEndDate),
		errors.Is(err, validators.ErrEndDateBeforeStart),
		errors.Is(err, errorsService.ErrAdminNoTerminal):
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	default:
		return echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
	}
}

func (h *NotificationHandler) NotifyBusDelay(c echo.Context) error {
	userID, ok := c.Get("user_id").(uuid.UUID)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid user context")
	}
	role, _ := c.Get("role").(string)

	var req models.NotifyBusDelayRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	resp, err := h.svc.NotifyBusDelay(c.Request().Context(), userID, role, req)
	if err != nil {
		return mapNotifyBusDelayError(err)
	}

	return c.JSON(http.StatusCreated, resp)
}

func mapNotificationError(err error) error {
	switch {
	case errors.Is(err, validators.ErrLicensePatentEmpty),
		errors.Is(err, validators.ErrCodeEmpty),
		errors.Is(err, validators.ErrInvalidCode),
		errors.Is(err, validators.ErrNotificationTimeLifeInvalid):
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	case errors.Is(err, repository.ErrNotFound):
		return echo.NewHTTPError(http.StatusNotFound, "platform not found")
	case errors.Is(err, errorsService.ErrPlatformMissingTerminal):
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	case errors.Is(err, errorsService.ErrNotification):
		return echo.NewHTTPError(http.StatusBadGateway, "failed to notify passengers")
	default:
		return err
	}
}

func mapAdminLocalNotificationError(err error) error {
	switch {
	case errors.Is(err, validators.ErrNotificationTypeInvalid),
		errors.Is(err, validators.ErrNotificationMessageEmpty),
		errors.Is(err, errorsService.ErrTerminalUUIDRequired),
		errors.Is(err, errorsService.ErrTerminalUUIDRequiredMultiAdmin),
		errors.Is(err, errorsService.ErrInvalidTerminalUUID),
		errors.Is(err, validators.ErrNotificationPayloadInvalidJSON),
		errors.Is(err, validators.ErrNotificationPayloadEmpty),
		errors.Is(err, validators.ErrNotificationTimeLifeInvalid):
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	case errors.Is(err, validators.ErrNotificationGlobalSuperAdminOnly):
		return echo.NewHTTPError(http.StatusForbidden, err.Error())
	case errors.Is(err, errorsService.ErrTerminalNotOwned):
		return echo.NewHTTPError(http.StatusForbidden, err.Error())
	case errors.Is(err, errorsService.ErrTerminalNotFound):
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	case errors.Is(err, errorsService.ErrAdminNoTerminal):
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	case errors.Is(err, errorsService.ErrNotification):
		return echo.NewHTTPError(http.StatusBadGateway, "failed to send notification")
	default:
		return err
	}
}

func mapCameraErrorNotifyError(err error) error {
	switch {
	case errors.Is(err, validators.ErrCameraNotificationTypeInvalid),
		errors.Is(err, validators.ErrCodeCameraEmpty),
		errors.Is(err, validators.ErrCodeCameraInvalid),
		errors.Is(err, validators.ErrCameraErrorMessageEmpty),
		errors.Is(err, validators.ErrNotificationTimeLifeInvalid):
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	case errors.Is(err, repository.ErrNotFound):
		return echo.NewHTTPError(http.StatusNotFound, "platform not found")
	case errors.Is(err, errorsService.ErrPlatformMissingTerminal):
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	case errors.Is(err, errorsService.ErrNotification):
		return echo.NewHTTPError(http.StatusBadGateway, "failed to notify admins")
	default:
		return err
	}
}

func (h *NotificationHandler) DeleteNotification(c echo.Context) error {
	userID, ok := c.Get("user_id").(uuid.UUID)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid user context")
	}
	role, _ := c.Get("role").(string)

	rawID := c.QueryParam("notification_id")
	if rawID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "notification_id is required")
	}
	notificationID, err := uuid.Parse(rawID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "notification_id must be a valid UUID")
	}

	if err := h.svc.DeleteNotification(c.Request().Context(), userID, role, notificationID); err != nil {
		return mapDeleteNotificationError(err)
	}
	return c.NoContent(http.StatusNoContent)
}

func mapDeleteNotificationError(err error) error {
	switch {
	case errors.Is(err, errorsService.ErrUserCannotDeleteNotification),
		errors.Is(err, errorsService.ErrNotificationDeleteForbidden):
		return echo.NewHTTPError(http.StatusForbidden, err.Error())
	case errors.Is(err, errorsService.ErrNotificationNotFound):
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	case errors.Is(err, errorsService.ErrAdminNoTerminal):
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	default:
		return echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
	}
}

func mapNotifyBusDelayError(err error) error {
	switch {
	case errors.Is(err, validators.ErrBusDelayTypeInvalid),
		errors.Is(err, validators.ErrBusDelayLicensePatentRequired),
		errors.Is(err, validators.ErrBusDelayStartDateRequired),
		errors.Is(err, validators.ErrBusDelayTimeDelayInvalid),
		errors.Is(err, errorsService.ErrBusDelayTerminalUUIDRequired),
		errors.Is(err, errorsService.ErrInvalidTerminalUUID),
		errors.Is(err, validators.ErrNotificationTimeLifeInvalid):
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	case errors.Is(err, errorsService.ErrTerminalNotOwned):
		return echo.NewHTTPError(http.StatusForbidden, err.Error())
	case errors.Is(err, errorsService.ErrTerminalNotFound):
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	case errors.Is(err, errorsService.ErrAdminNoTerminal):
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	case errors.Is(err, errorsService.ErrExternalTerminalNotConfigured),
		errors.Is(err, errorsService.ErrTripNotRegistered):
		return echo.NewHTTPError(http.StatusConflict, err.Error())
	case errors.Is(err, errorsService.ErrExternalTerminalIDRequired),
		errors.Is(err, errorsService.ErrUpstreamRequest),
		errors.Is(err, errorsService.ErrUpstreamResponse):
		return echo.NewHTTPError(http.StatusBadGateway, err.Error())
	case errors.Is(err, errorsService.ErrNotification):
		return echo.NewHTTPError(http.StatusBadGateway, "failed to send bus delay notification")
	default:
		return err
	}
}
