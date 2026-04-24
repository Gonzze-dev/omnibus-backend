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
		if errors.Is(err, service.ErrUnsupportedNotificationRole) {
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
	case errors.Is(err, service.ErrTerminalIDRequired),
		errors.Is(err, service.ErrInvalidTerminalID),
		errors.Is(err, service.ErrNotificationTypeInvalid),
		errors.Is(err, service.ErrInvalidStartDate),
		errors.Is(err, service.ErrInvalidEndDate),
		errors.Is(err, service.ErrEndDateBeforeStart),
		errors.Is(err, service.ErrAdminNoTerminal):
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
	case errors.Is(err, service.ErrLicensePatentEmpty),
		errors.Is(err, service.ErrCodeEmpty),
		errors.Is(err, service.ErrInvalidCode),
		errors.Is(err, service.ErrNotificationTimeLifeInvalid):
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	case errors.Is(err, repository.ErrNotFound):
		return echo.NewHTTPError(http.StatusNotFound, "platform not found")
	case errors.Is(err, service.ErrPlatformMissingTerminal):
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	case errors.Is(err, service.ErrNotification):
		return echo.NewHTTPError(http.StatusBadGateway, "failed to notify passengers")
	default:
		return err
	}
}

func mapAdminLocalNotificationError(err error) error {
	switch {
	case errors.Is(err, service.ErrNotificationTypeInvalid),
		errors.Is(err, service.ErrNotificationMessageEmpty),
		errors.Is(err, service.ErrTerminalUUIDRequired),
		errors.Is(err, service.ErrTerminalUUIDRequiredMultiAdmin),
		errors.Is(err, service.ErrInvalidTerminalUUID),
		errors.Is(err, service.ErrNotificationPayloadInvalidJSON),
		errors.Is(err, service.ErrNotificationPayloadEmpty),
		errors.Is(err, service.ErrNotificationTimeLifeInvalid):
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	case errors.Is(err, service.ErrNotificationGlobalSuperAdminOnly):
		return echo.NewHTTPError(http.StatusForbidden, err.Error())
	case errors.Is(err, service.ErrTerminalNotOwned):
		return echo.NewHTTPError(http.StatusForbidden, err.Error())
	case errors.Is(err, service.ErrTerminalNotFound):
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	case errors.Is(err, service.ErrAdminNoTerminal):
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	case errors.Is(err, service.ErrNotification):
		return echo.NewHTTPError(http.StatusBadGateway, "failed to send notification")
	default:
		return err
	}
}

func mapCameraErrorNotifyError(err error) error {
	switch {
	case errors.Is(err, service.ErrCameraNotificationTypeInvalid),
		errors.Is(err, service.ErrCodeCameraEmpty),
		errors.Is(err, service.ErrCodeCameraInvalid),
		errors.Is(err, service.ErrCameraErrorMessageEmpty),
		errors.Is(err, service.ErrNotificationTimeLifeInvalid):
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	case errors.Is(err, repository.ErrNotFound):
		return echo.NewHTTPError(http.StatusNotFound, "platform not found")
	case errors.Is(err, service.ErrPlatformMissingTerminal):
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	case errors.Is(err, service.ErrNotification):
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
	case errors.Is(err, service.ErrUserCannotDeleteNotification),
		errors.Is(err, service.ErrNotificationDeleteForbidden):
		return echo.NewHTTPError(http.StatusForbidden, err.Error())
	case errors.Is(err, service.ErrNotificationNotFound):
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	case errors.Is(err, service.ErrAdminNoTerminal):
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	default:
		return echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
	}
}

func mapNotifyBusDelayError(err error) error {
	switch {
	case errors.Is(err, service.ErrBusDelayTypeInvalid),
		errors.Is(err, service.ErrBusDelayLicensePatentRequired),
		errors.Is(err, service.ErrBusDelayStartDateRequired),
		errors.Is(err, service.ErrBusDelayTimeDelayInvalid),
		errors.Is(err, service.ErrBusDelayTerminalUUIDRequired),
		errors.Is(err, service.ErrInvalidTerminalUUID),
		errors.Is(err, service.ErrNotificationTimeLifeInvalid):
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	case errors.Is(err, service.ErrTerminalNotOwned):
		return echo.NewHTTPError(http.StatusForbidden, err.Error())
	case errors.Is(err, service.ErrTerminalNotFound):
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	case errors.Is(err, service.ErrAdminNoTerminal):
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	case errors.Is(err, service.ErrExternalTerminalNotConfigured),
		errors.Is(err, service.ErrTripNotRegistered):
		return echo.NewHTTPError(http.StatusConflict, err.Error())
	case errors.Is(err, service.ErrExternalTerminalIDRequired),
		errors.Is(err, service.ErrUpstreamRequest),
		errors.Is(err, service.ErrUpstreamResponse):
		return echo.NewHTTPError(http.StatusBadGateway, err.Error())
	case errors.Is(err, service.ErrNotification):
		return echo.NewHTTPError(http.StatusBadGateway, "failed to send bus delay notification")
	default:
		return err
	}
}
