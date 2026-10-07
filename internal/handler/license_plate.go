package handler

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"

	errorsService "tesina/backend/internal/errors"
	"tesina/backend/internal/service"
)

// maxLicensePlateImageSize limita el tamaño de la foto que se reenvia al OCR.
const maxLicensePlateImageSize = 10 << 20

type LicensePlateHandler struct {
	svc service.LicensePlateService
}

func NewLicensePlateHandler(svc service.LicensePlateService) *LicensePlateHandler {
	return &LicensePlateHandler{svc: svc}
}

type readLicensePlateResponse struct {
	LicensePlate string `json:"license_plate"`
}

// Read recibe una foto (multipart, campo "image") y devuelve la patente leida.
func (h *LicensePlateHandler) Read(c echo.Context) error {
	fileHeader, err := c.FormFile("image")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errorsService.ErrLicensePlateImageRequired.Error())
	}
	if fileHeader.Size > maxLicensePlateImageSize {
		return echo.NewHTTPError(http.StatusRequestEntityTooLarge, "image must be at most 10MB")
	}

	file, err := fileHeader.Open()
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errorsService.ErrLicensePlateImageInvalid.Error())
	}
	defer file.Close()

	plate, err := h.svc.Read(c.Request().Context(), fileHeader.Filename, file)
	if err != nil {
		return mapLicensePlateError(err)
	}
	return c.JSON(http.StatusOK, readLicensePlateResponse{LicensePlate: plate})
}

func mapLicensePlateError(err error) error {
	switch {
	case errors.Is(err, errorsService.ErrLicensePlateImageRequired),
		errors.Is(err, errorsService.ErrLicensePlateImageInvalid):
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	case errors.Is(err, errorsService.ErrLicensePlateNotRead):
		return echo.NewHTTPError(http.StatusUnprocessableEntity, err.Error())
	case errors.Is(err, errorsService.ErrUpstreamRequest),
		errors.Is(err, errorsService.ErrUpstreamResponse):
		return echo.NewHTTPError(http.StatusBadGateway, "license plate reader unavailable")
	default:
		return echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
	}
}
