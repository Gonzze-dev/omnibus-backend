package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	errorsService "tesina/backend/internal/errors"
	"tesina/backend/internal/models"
	"tesina/backend/internal/roles"
	"tesina/backend/internal/service"
	"tesina/backend/internal/validators"
)

type AdminHandler struct {
	svc service.AdminService
}

func NewAdminHandler(svc service.AdminService) *AdminHandler {
	return &AdminHandler{svc: svc}
}

// --- Cities ---

func (h *AdminHandler) ListCities(c echo.Context) error {
	params := models.ListCitiesParams{
		Page:  1,
		Limit: 10,
		Order: "DESC",
	}
	if raw := c.QueryParam("page"); raw != "" {
		if v, err := strconv.Atoi(raw); err == nil && v > 0 {
			params.Page = v
		}
	}
	if raw := c.QueryParam("limit"); raw != "" {
		if v, err := strconv.Atoi(raw); err == nil && v > 0 {
			params.Limit = v
		}
	}
	if raw := c.QueryParam("order"); raw != "" {
		params.Order = raw
	}

	res, err := h.svc.ListCities(c.Request().Context(), params)
	if err != nil {
		return mapAdminError(err)
	}
	return c.JSON(http.StatusOK, res)
}

func (h *AdminHandler) CountCities(c echo.Context) error {
	total, err := h.svc.CountCities(c.Request().Context())
	if err != nil {
		return mapAdminError(err)
	}
	return c.JSON(http.StatusOK, models.CountCitiesResponse{Total: total})
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
	busTerminalID, err := platformTerminalFilter(c)
	if err != nil {
		return err
	}

	params := models.ListPlatformsParams{
		Page:  1,
		Limit: 10,
		Order: "DESC",
	}
	if raw := c.QueryParam("page"); raw != "" {
		if v, err := strconv.Atoi(raw); err == nil && v > 0 {
			params.Page = v
		}
	}
	if raw := c.QueryParam("limit"); raw != "" {
		if v, err := strconv.Atoi(raw); err == nil && v > 0 {
			params.Limit = v
		}
	}
	if raw := c.QueryParam("order"); raw != "" {
		params.Order = raw
	}

	role, _ := c.Get("role").(string)
	if role == roles.SuperAdmin {
		res, err := h.svc.ListAllPlatforms(c.Request().Context(), busTerminalID, params)
		if err != nil {
			return mapAdminError(err)
		}
		return c.JSON(http.StatusOK, res)
	}

	adminID, err := getUserID(c)
	if err != nil {
		return err
	}

	res, err := h.svc.ListPlatforms(c.Request().Context(), adminID, busTerminalID, params)
	if err != nil {
		return mapAdminError(err)
	}
	return c.JSON(http.StatusOK, res)
}

func (h *AdminHandler) CountPlatforms(c echo.Context) error {
	busTerminalID, err := platformTerminalFilter(c)
	if err != nil {
		return err
	}

	role, _ := c.Get("role").(string)
	if role == roles.SuperAdmin {
		total, err := h.svc.CountAllPlatforms(c.Request().Context(), busTerminalID)
		if err != nil {
			return mapAdminError(err)
		}
		return c.JSON(http.StatusOK, models.CountPlatformsResponse{Total: total})
	}

	adminID, err := getUserID(c)
	if err != nil {
		return err
	}

	total, err := h.svc.CountPlatforms(c.Request().Context(), adminID, busTerminalID)
	if err != nil {
		return mapAdminError(err)
	}
	return c.JSON(http.StatusOK, models.CountPlatformsResponse{Total: total})
}

// platformTerminalFilter lee el query param opcional bus_terminal_id.
func platformTerminalFilter(c echo.Context) (*uuid.UUID, error) {
	raw := c.QueryParam("bus_terminal_id")
	if raw == "" {
		return nil, nil
	}
	id, err := uuid.Parse(raw)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "invalid bus_terminal_id")
	}
	return &id, nil
}

func (h *AdminHandler) GetPlatform(c echo.Context) error {
	code, err := strconv.Atoi(c.Param("code"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid platform code")
	}

	role, _ := c.Get("role").(string)
	if role == roles.SuperAdmin {
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
	if role == roles.SuperAdmin {
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
	if role == roles.SuperAdmin {
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
	if role == roles.SuperAdmin {
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

func (h *AdminHandler) ListUsers(c echo.Context) error {
	params := models.ListUsersParams{
		Page:  1,
		Limit: 10,
		Order: "DESC",
	}
	if raw := c.QueryParam("page"); raw != "" {
		if v, err := strconv.Atoi(raw); err == nil && v > 0 {
			params.Page = v
		}
	}
	if raw := c.QueryParam("limit"); raw != "" {
		if v, err := strconv.Atoi(raw); err == nil && v > 0 {
			params.Limit = v
		}
	}
	if raw := c.QueryParam("order"); raw != "" {
		params.Order = raw
	}
	params.Search = c.QueryParam("search")

	res, err := h.svc.ListUsers(c.Request().Context(), params)
	if err != nil {
		return mapAdminError(err)
	}
	return c.JSON(http.StatusOK, res)
}

func (h *AdminHandler) CountUsers(c echo.Context) error {
	total, err := h.svc.CountUsers(c.Request().Context())
	if err != nil {
		return mapAdminError(err)
	}
	return c.JSON(http.StatusOK, models.CountUsersResponse{Total: total})
}

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
	if role == roles.SuperAdmin {
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
	if role == roles.SuperAdmin {
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

// --- Stats ---

func (h *AdminHandler) CountActiveTerminals(c echo.Context) error {
	total, err := h.svc.CountActiveTerminals(c.Request().Context())
	if err != nil {
		return mapAdminError(err)
	}
	return c.JSON(http.StatusOK, models.CountActiveResponse{Total: total})
}

func (h *AdminHandler) CountActivePlatforms(c echo.Context) error {
	total, err := h.svc.CountActivePlatforms(c.Request().Context())
	if err != nil {
		return mapAdminError(err)
	}
	return c.JSON(http.StatusOK, models.CountActiveResponse{Total: total})
}

func (h *AdminHandler) CountActiveCities(c echo.Context) error {
	total, err := h.svc.CountActiveCities(c.Request().Context())
	if err != nil {
		return mapAdminError(err)
	}
	return c.JSON(http.StatusOK, models.CountActiveResponse{Total: total})
}

func mapAdminError(err error) error {
	switch {
	case errors.Is(err, validators.ErrPostalCodeRequired),
		errors.Is(err, validators.ErrCityNameRequired),
		errors.Is(err, validators.ErrAndenRequired),
		errors.Is(err, validators.ErrEmailRequired):
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	case errors.Is(err, errorsService.ErrCityNotFound):
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	case errors.Is(err, errorsService.ErrCityAlreadyExists):
		return echo.NewHTTPError(http.StatusConflict, err.Error())
	case errors.Is(err, errorsService.ErrPlatformNotFound):
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	case errors.Is(err, errorsService.ErrTerminalNotFound):
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	case errors.Is(err, errorsService.ErrTerminalNotOwned):
		return echo.NewHTTPError(http.StatusForbidden, err.Error())
	case errors.Is(err, errorsService.ErrUserNotFound):
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	case errors.Is(err, errorsService.ErrAlreadyAdmin):
		return echo.NewHTTPError(http.StatusConflict, err.Error())
	case errors.Is(err, errorsService.ErrAlreadySuperAdmin):
		return echo.NewHTTPError(http.StatusConflict, err.Error())
	case errors.Is(err, errorsService.ErrNotAdmin):
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	case errors.Is(err, errorsService.ErrCannotDemoteSelf):
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	default:
		return echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
	}
}
