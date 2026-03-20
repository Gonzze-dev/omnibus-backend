package handler

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"

	"tesina/backend/internal/models"
	"tesina/backend/internal/service"
)

const (
	refreshTokenCookie = "refresh_token"
	refreshTokenMaxAge = 7 * 24 * 60 * 60 // 7 days in seconds
)

type AuthHandler struct {
	svc service.AuthService
}

func NewAuthHandler(svc service.AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

func (h *AuthHandler) Register(c echo.Context) error {
	var req models.CreateUserRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	resp, err := h.svc.Register(c.Request().Context(), req)
	if err != nil {
		return mapAuthError(err)
	}

	return c.JSON(http.StatusCreated, resp)
}

func (h *AuthHandler) Login(c echo.Context) error {
	var req models.LoginRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	resp, err := h.svc.Login(c.Request().Context(), req)
	if err != nil {
		return mapAuthError(err)
	}

	setRefreshTokenCookie(c, resp.RefreshToken)
	return c.JSON(http.StatusOK, resp)
}

func (h *AuthHandler) RefreshToken(c echo.Context) error {
	token, err := readRefreshTokenCookie(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "refresh token cookie missing")
	}

	req := models.RefreshTokenRequest{RefreshToken: token}
	resp, err := h.svc.RefreshToken(c.Request().Context(), req)
	if err != nil {
		return mapAuthError(err)
	}

	setRefreshTokenCookie(c, resp.RefreshToken)
	return c.JSON(http.StatusOK, resp)
}

func (h *AuthHandler) Logout(c echo.Context) error {
	token, err := readRefreshTokenCookie(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "refresh token cookie missing")
	}

	req := models.LogoutRequest{RefreshToken: token}
	if err := h.svc.Logout(c.Request().Context(), req); err != nil {
		return mapAuthError(err)
	}

	clearRefreshTokenCookie(c)
	return c.JSON(http.StatusOK, map[string]string{"message": "logged out successfully"})
}

func setRefreshTokenCookie(c echo.Context, token string) {
	c.SetCookie(&http.Cookie{
		Name:     refreshTokenCookie,
		Value:    token,
		Path:     "/api/auth",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   refreshTokenMaxAge,
	})
}

func readRefreshTokenCookie(c echo.Context) (string, error) {
	cookie, err := c.Cookie(refreshTokenCookie)
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

func clearRefreshTokenCookie(c echo.Context) {
	c.SetCookie(&http.Cookie{
		Name:     refreshTokenCookie,
		Value:    "",
		Path:     "/api/auth",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
}

func mapAuthError(err error) error {
	switch {
	case errors.Is(err, service.ErrMissingFields):
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	case errors.Is(err, service.ErrEmailAlreadyExists):
		return echo.NewHTTPError(http.StatusConflict, err.Error())
	case errors.Is(err, service.ErrDNIAlreadyExists):
		return echo.NewHTTPError(http.StatusConflict, err.Error())
	case errors.Is(err, service.ErrInvalidCredentials):
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
	case errors.Is(err, service.ErrInvalidRefreshToken):
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
	default:
		return echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
	}
}
