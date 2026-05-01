package handler

import (
	"errors"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"tesina/backend/internal/models"
	"tesina/backend/internal/service"
	"tesina/backend/internal/validators"
	errorsService "tesina/backend/internal/errors"
)

type PasswordRecoveryHandler struct {
	svc service.PasswordRecoveryService
}

func NewPasswordRecoveryHandler(svc service.PasswordRecoveryService) *PasswordRecoveryHandler {
	return &PasswordRecoveryHandler{svc: svc}
}

func (h *PasswordRecoveryHandler) ForgotPassword(c echo.Context) error {
	var req models.ForgotPasswordRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	resp, err := h.svc.ForgotPassword(c.Request().Context(), req.Email)
	if err != nil {
		return mapPasswordRecoveryError(err)
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *PasswordRecoveryHandler) ValidateRecoveryToken(c echo.Context) error {
	token, err := bearerTokenFromRequest(c)
	if err != nil {
		return err
	}

	resp, err := h.svc.ValidateRecoveryToken(c.Request().Context(), token)
	if err != nil {
		return mapPasswordRecoveryError(err)
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *PasswordRecoveryHandler) ResetPassword(c echo.Context) error {
	token, err := bearerTokenFromRequest(c)
	if err != nil {
		return err
	}

	var req models.ResetPasswordRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := h.svc.ResetPasswordWithToken(c.Request().Context(), token, req.Password); err != nil {
		return mapPasswordRecoveryError(err)
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "password updated"})
}

func bearerTokenFromRequest(c echo.Context) (string, error) {
	authHeader := c.Request().Header.Get("Authorization")
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || strings.TrimSpace(parts[1]) == "" {
		return "", echo.NewHTTPError(http.StatusUnauthorized, "missing or invalid authorization bearer token")
	}
	return strings.TrimSpace(parts[1]), nil
}

func mapPasswordRecoveryError(err error) error {
	switch {
	case errors.Is(err, validators.ErrEmailRequired),
		errors.Is(err, validators.ErrPasswordRequired),
		errors.Is(err, validators.ErrTokenRequired):
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	case errors.Is(err, errorsService.ErrInvalidPasswordResetToken):
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid or expired token")
	default:
		return echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
	}
}
