package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"tesina/backend/internal/models"
	"tesina/backend/internal/service"
)

type AdminHandler struct {
	svc service.AdminService
}

func NewAdminHandler(svc service.AdminService) *AdminHandler {
	return &AdminHandler{svc: svc}
}

// --- Cities ---

func (h *AdminHandler) ListCities(c echo.Context) error {
	cities, err := h.svc.ListCities(c.Request().Context())
	if err != nil {
		return mapAdminError(err)
	}
	return c.JSON(http.StatusOK, cities)
}

func (h *AdminHandler) GetCity(c echo.Context) error {
	postalCode := c.Param("postal_code")
	city, err := h.svc.GetCity(c.Request().Context(), postalCode)
	if err != nil {
		return mapAdminError(err)
	}
	return c.JSON(http.StatusOK, city)
}

func (h *AdminHandler) CreateCity(c echo.Context) error {
	var req models.CreateCityRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	city, err := h.svc.CreateCity(c.Request().Context(), req)
	if err != nil {
		return mapAdminError(err)
	}
	return c.JSON(http.StatusCreated, city)
}

func (h *AdminHandler) UpdateCity(c echo.Context) error {
	postalCode := c.Param("postal_code")
	var req models.UpdateCityRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	city, err := h.svc.UpdateCity(c.Request().Context(), postalCode, req)
	if err != nil {
		return mapAdminError(err)
	}
	return c.JSON(http.StatusOK, city)
}

func (h *AdminHandler) DeleteCity(c echo.Context) error {
	postalCode := c.Param("postal_code")
	if err := h.svc.DeleteCity(c.Request().Context(), postalCode); err != nil {
		return mapAdminError(err)
	}
	return c.NoContent(http.StatusNoContent)
}

// --- Platforms ---

func (h *AdminHandler) ListPlatforms(c echo.Context) error {
	var busTerminalID *uuid.UUID
	if raw := c.QueryParam("bus_terminal_id"); raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid bus_terminal_id")
		}
		busTerminalID = &id
	}

	role, _ := c.Get("role").(string)
	if role == "super_admin" {
		terminals, err := h.svc.ListAllPlatforms(c.Request().Context(), busTerminalID)
		if err != nil {
			return mapAdminError(err)
		}
		return c.JSON(http.StatusOK, terminals)
	}

	adminID, err := getUserID(c)
	if err != nil {
		return err
	}

	terminals, err := h.svc.ListPlatforms(c.Request().Context(), adminID, busTerminalID)
	if err != nil {
		return mapAdminError(err)
	}
	return c.JSON(http.StatusOK, terminals)
}

func (h *AdminHandler) GetPlatform(c echo.Context) error {
	code, err := strconv.Atoi(c.Param("code"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid platform code")
	}

	role, _ := c.Get("role").(string)
	if role == "super_admin" {
		platform, err := h.svc.GetPlatformByCode(c.Request().Context(), code)
		if err != nil {
			return mapAdminError(err)
		}
		return c.JSON(http.StatusOK, platform)
	}

	adminID, err := getUserID(c)
	if err != nil {
		return err
	}

	platform, err := h.svc.GetPlatform(c.Request().Context(), adminID, code)
	if err != nil {
		return mapAdminError(err)
	}
	return c.JSON(http.StatusOK, platform)
}

func (h *AdminHandler) CreatePlatform(c echo.Context) error {
	var req models.CreatePlatformRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	role, _ := c.Get("role").(string)
	if role == "super_admin" {
		platform, err := h.svc.CreatePlatformDirect(c.Request().Context(), req)
		if err != nil {
			return mapAdminError(err)
		}
		return c.JSON(http.StatusCreated, platform)
	}

	adminID, err := getUserID(c)
	if err != nil {
		return err
	}

	platform, err := h.svc.CreatePlatform(c.Request().Context(), adminID, req)
	if err != nil {
		return mapAdminError(err)
	}
	return c.JSON(http.StatusCreated, platform)
}

func (h *AdminHandler) UpdatePlatform(c echo.Context) error {
	code, err := strconv.Atoi(c.Param("code"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid platform code")
	}

	var req models.UpdatePlatformRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	role, _ := c.Get("role").(string)
	if role == "super_admin" {
		platform, err := h.svc.UpdatePlatformByCode(c.Request().Context(), code, req)
		if err != nil {
			return mapAdminError(err)
		}
		return c.JSON(http.StatusOK, platform)
	}

	adminID, err := getUserID(c)
	if err != nil {
		return err
	}

	platform, err := h.svc.UpdatePlatform(c.Request().Context(), adminID, code, req)
	if err != nil {
		return mapAdminError(err)
	}
	return c.JSON(http.StatusOK, platform)
}

func (h *AdminHandler) DeletePlatform(c echo.Context) error {
	code, err := strconv.Atoi(c.Param("code"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid platform code")
	}

	role, _ := c.Get("role").(string)
	if role == "super_admin" {
		if err := h.svc.DeletePlatformByCode(c.Request().Context(), code); err != nil {
			return mapAdminError(err)
		}
		return c.NoContent(http.StatusNoContent)
	}

	adminID, err := getUserID(c)
	if err != nil {
		return err
	}

	if err := h.svc.DeletePlatform(c.Request().Context(), adminID, code); err != nil {
		return mapAdminError(err)
	}
	return c.NoContent(http.StatusNoContent)
}

// --- User management ---

func (h *AdminHandler) GetUserByEmail(c echo.Context) error {
	email := c.QueryParam("email")
	resp, err := h.svc.GetUserByEmail(c.Request().Context(), email)
	if err != nil {
		return mapAdminError(err)
	}
	return c.JSON(http.StatusOK, resp)
}

func (h *AdminHandler) PromoteToAdmin(c echo.Context) error {
	var req models.PromoteAdminRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	role, _ := c.Get("role").(string)
	if role == "super_admin" {
		resp, err := h.svc.PromoteToAdminDirect(c.Request().Context(), req)
		if err != nil {
			return mapAdminError(err)
		}
		return c.JSON(http.StatusOK, resp)
	}

	adminID, err := getUserID(c)
	if err != nil {
		return err
	}

	resp, err := h.svc.PromoteToAdmin(c.Request().Context(), adminID, req)
	if err != nil {
		return mapAdminError(err)
	}
	return c.JSON(http.StatusOK, resp)
}

func (h *AdminHandler) DemoteAdmin(c echo.Context) error {
	var req models.DemoteAdminRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	role, _ := c.Get("role").(string)
	if role == "super_admin" {
		resp, err := h.svc.DemoteAdminDirect(c.Request().Context(), req)
		if err != nil {
			return mapAdminError(err)
		}
		return c.JSON(http.StatusOK, resp)
	}

	adminID, err := getUserID(c)
	if err != nil {
		return err
	}

	resp, err := h.svc.DemoteAdmin(c.Request().Context(), adminID, req)
	if err != nil {
		return mapAdminError(err)
	}
	return c.JSON(http.StatusOK, resp)
}

func mapAdminError(err error) error {
	switch {
	case errors.Is(err, service.ErrMissingFields):
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	case errors.Is(err, service.ErrCityNotFound):
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	case errors.Is(err, service.ErrCityAlreadyExists):
		return echo.NewHTTPError(http.StatusConflict, err.Error())
	case errors.Is(err, service.ErrPlatformNotFound):
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	case errors.Is(err, service.ErrTerminalNotFound):
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	case errors.Is(err, service.ErrTerminalNotOwned):
		return echo.NewHTTPError(http.StatusForbidden, err.Error())
	case errors.Is(err, service.ErrUserNotFound):
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	case errors.Is(err, service.ErrAlreadyAdmin):
		return echo.NewHTTPError(http.StatusConflict, err.Error())
	case errors.Is(err, service.ErrAlreadySuperAdmin):
		return echo.NewHTTPError(http.StatusConflict, err.Error())
	case errors.Is(err, service.ErrNotAdmin):
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	case errors.Is(err, service.ErrCannotDemoteSelf):
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	default:
		return echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
	}
}
