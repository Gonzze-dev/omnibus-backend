package handler

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"

	"tesina/backend/internal/models"
	"tesina/backend/internal/service"
	errorsService "tesina/backend/internal/errors"
)

type BusHandler struct {
	svc service.BusService
}

func NewBusHandler(svc service.BusService) *BusHandler {
	return &BusHandler{svc: svc}
}

func (h *BusHandler) JoinBus(c echo.Context) error {
	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	var req models.JoinBusRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := h.svc.JoinBus(c.Request().Context(), userID, req); err != nil {
		return mapJoinBusError(err)
	}

	return c.NoContent(http.StatusCreated)
}

func mapJoinBusError(err error) error {
	switch {
	case errors.Is(err, errorsService.ErrTerminalIDRequired),
		errors.Is(err, errorsService.ErrTerminalIDInvalid),
		errors.Is(err, errorsService.ErrTicketRequired):
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	case errors.Is(err, errorsService.ErrTerminalNotFound),
		errors.Is(err, errorsService.ErrTripNotFound):
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	default:
		return echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
	}
}
