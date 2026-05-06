package handler

import (
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"tesina/backend/internal/models"
	"tesina/backend/internal/service"
	"tesina/backend/internal/validators"
	errorsService "tesina/backend/internal/errors"
)

type UserHandler struct {
	svc service.UserService
}

func NewUserHandler(svc service.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

func (h *UserHandler) GetProfile(c echo.Context) error {
	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	resp, err := h.svc.GetProfile(c.Request().Context(), userID)
	if err != nil {
		return mapUserError(err)
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *UserHandler) UpdateProfile(c echo.Context) error {
	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	var req models.UpdateUserRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	resp, err := h.svc.UpdateProfile(c.Request().Context(), userID, req)
	if err != nil {
		return mapUserError(err)
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *UserHandler) ListTerminals(c echo.Context) error {
	terminals, err := h.svc.ListTerminals(c.Request().Context())
	if err != nil {
		return mapUserError(err)
	}
	return c.JSON(http.StatusOK, terminals)
}

func (h *UserHandler) DeleteAccount(c echo.Context) error {
	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	if err := h.svc.DeleteAccount(c.Request().Context(), userID); err != nil {
		return mapUserError(err)
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "account deleted successfully"})
}

func getUserID(c echo.Context) (uuid.UUID, error) {
	id, ok := c.Get("user_id").(uuid.UUID)
	if !ok {
		return uuid.Nil, echo.NewHTTPError(http.StatusUnauthorized, "user id not found in context")
	}
	return id, nil
}

func mapUserError(err error) error {
	switch {
	case errors.Is(err, errorsService.ErrUserNotFound):
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	case errors.Is(err, errorsService.ErrEmailAlreadyExists):
		return echo.NewHTTPError(http.StatusConflict, err.Error())
	case errors.Is(err, errorsService.ErrMissingFields),
		errors.Is(err, validators.ErrNoFieldsToUpdate),
		errors.Is(err, validators.ErrFirstNameEmpty),
		errors.Is(err, validators.ErrLastNameEmpty),
		errors.Is(err, validators.ErrEmailEmpty),
		errors.Is(err, validators.ErrPasswordEmpty):
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	default:
		return echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
	}
}
