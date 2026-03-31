package handler

import (
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"tesina/backend/internal/models"
	"tesina/backend/internal/service"
)

type SuperAdminHandler struct {
	svc service.SuperAdminService
}

func NewSuperAdminHandler(svc service.SuperAdminService) *SuperAdminHandler {
	return &SuperAdminHandler{svc: svc}
}

// --- Terminals ---

func (h *SuperAdminHandler) ListTerminals(c echo.Context) error {
	terminals, err := h.svc.ListTerminals(c.Request().Context())
	if err != nil {
		return mapSuperAdminError(err)
	}
	return c.JSON(http.StatusOK, terminals)
}

func (h *SuperAdminHandler) GetTerminal(c echo.Context) error {
	id, err := uuid.Parse(c.Param("uuid"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid terminal uuid")
	}

	terminal, err := h.svc.GetTerminal(c.Request().Context(), id)
	if err != nil {
		return mapSuperAdminError(err)
	}
	return c.JSON(http.StatusOK, terminal)
}

func (h *SuperAdminHandler) CreateTerminal(c echo.Context) error {
	var req models.CreateBusTerminalRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	terminal, err := h.svc.CreateTerminal(c.Request().Context(), req)
	if err != nil {
		return mapSuperAdminError(err)
	}
	return c.JSON(http.StatusCreated, terminal)
}

func (h *SuperAdminHandler) UpdateTerminal(c echo.Context) error {
	id, err := uuid.Parse(c.Param("uuid"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid terminal uuid")
	}

	var req models.UpdateBusTerminalRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	terminal, err := h.svc.UpdateTerminal(c.Request().Context(), id, req)
	if err != nil {
		return mapSuperAdminError(err)
	}
	return c.JSON(http.StatusOK, terminal)
}

func (h *SuperAdminHandler) DeleteTerminal(c echo.Context) error {
	id, err := uuid.Parse(c.Param("uuid"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid terminal uuid")
	}

	if err := h.svc.DeleteTerminal(c.Request().Context(), id); err != nil {
		return mapSuperAdminError(err)
	}
	return c.NoContent(http.StatusNoContent)
}

// --- User management (super-only) ---

func (h *SuperAdminHandler) PromoteToSuper(c echo.Context) error {
	var req models.PromoteSuperRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	resp, err := h.svc.PromoteToSuper(c.Request().Context(), req)
	if err != nil {
		return mapSuperAdminError(err)
	}
	return c.JSON(http.StatusOK, resp)
}

func (h *SuperAdminHandler) DemoteSuper(c echo.Context) error {
	var req models.DemoteSuperRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	resp, err := h.svc.DemoteSuper(c.Request().Context(), req)
	if err != nil {
		return mapSuperAdminError(err)
	}
	return c.JSON(http.StatusOK, resp)
}

func mapSuperAdminError(err error) error {
	switch {
	case errors.Is(err, service.ErrMissingFields):
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	case errors.Is(err, service.ErrExternalTerminalIDRequired):
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	case errors.Is(err, service.ErrInvalidExternalTerminalID):
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	case errors.Is(err, service.ErrExternalTerminalIDAlreadyUsed):
		return echo.NewHTTPError(http.StatusConflict, err.Error())
	case errors.Is(err, service.ErrUpstreamRequest), errors.Is(err, service.ErrUpstreamResponse):
		return echo.NewHTTPError(http.StatusBadGateway, err.Error())
	case errors.Is(err, service.ErrTerminalNotFound):
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	case errors.Is(err, service.ErrCityNotFound):
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	case errors.Is(err, service.ErrUserNotFound):
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	case errors.Is(err, service.ErrAlreadySuperAdmin):
		return echo.NewHTTPError(http.StatusConflict, err.Error())
	case errors.Is(err, service.ErrNotSuperAdmin):
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	default:
		return echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
	}
}
