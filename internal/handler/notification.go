package handler

import (
	"errors"
	"net/http"

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
		errors.Is(err, service.ErrInvalidCode):
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
		errors.Is(err, service.ErrNotificationPayloadEmpty):
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
		errors.Is(err, service.ErrCameraErrorMessageEmpty):
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

func mapNotifyBusDelayError(err error) error {
	switch {
	case errors.Is(err, service.ErrBusDelayTypeInvalid),
		errors.Is(err, service.ErrBusDelayLicensePatentRequired),
		errors.Is(err, service.ErrBusDelayStartDateRequired),
		errors.Is(err, service.ErrBusDelayTimeDelayInvalid),
		errors.Is(err, service.ErrBusDelayTerminalUUIDRequired),
		errors.Is(err, service.ErrInvalidTerminalUUID):
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
