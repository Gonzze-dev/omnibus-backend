package handler

import (
	"errors"
	"net/http"

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

func mapNotificationError(err error) error {
	switch {
	case errors.Is(err, service.ErrLicensePatentEmpty),
		errors.Is(err, service.ErrCodeEmpty),
		errors.Is(err, service.ErrInvalidCode):
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	case errors.Is(err, repository.ErrNotFound):
		return echo.NewHTTPError(http.StatusNotFound, "platform not found")
	case errors.Is(err, service.ErrNotification):
		return echo.NewHTTPError(http.StatusBadGateway, "failed to notify passengers")
	default:
		return err
	}
}
