package handler

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"

	errorsService "tesina/backend/internal/errors"
	"tesina/backend/internal/models"
	"tesina/backend/internal/service"
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

	resp, err := h.svc.JoinBus(c.Request().Context(), userID, req)
	if err != nil {
		return mapBusError(err)
	}

	return c.JSON(http.StatusCreated, resp)
}

func (h *BusHandler) GetAwaitedTrip(c echo.Context) error {
	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	resp, err := h.svc.GetAwaitedTrip(c.Request().Context(), userID)
	if err != nil {
		return mapBusError(err)
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *BusHandler) LeaveBus(c echo.Context) error {
	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	if err := h.svc.LeaveBus(c.Request().Context(), userID); err != nil {
		return mapBusError(err)
	}

	return c.NoContent(http.StatusNoContent)
}

func mapBusError(err error) error {
	switch {
	case errors.Is(err, errorsService.ErrTerminalIDRequired),
		errors.Is(err, errorsService.ErrTerminalIDInvalid),
		errors.Is(err, errorsService.ErrTicketRequired):
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	case errors.Is(err, errorsService.ErrTerminalNotFound),
		errors.Is(err, errorsService.ErrTripNotFound),
		errors.Is(err, errorsService.ErrAwaitedTripNotFound):
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	case errors.Is(err, errorsService.ErrTicketWrongTerminal):
		return echo.NewHTTPError(http.StatusConflict, err.Error())
	case errors.Is(err, errorsService.ErrUpstreamRequest),
		errors.Is(err, errorsService.ErrUpstreamResponse):
		return echo.NewHTTPError(http.StatusBadGateway, "terminal system unavailable")
	default:
		return echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
	}
}
