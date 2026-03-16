package handler

import (
	"encoding/json"
	"net/http"

	"github.com/labstack/echo/v4"

	"tesina/backend/internal/models"
	"tesina/backend/internal/service"
)

type PasajeHandler struct {
	svc service.PasajeService
}

func NewPasajeHandler(svc service.PasajeService) *PasajeHandler {
	return &PasajeHandler{svc: svc}
}

func (h *PasajeHandler) GetPasaje(c echo.Context) error {
	ticketString := c.Param("ticket_string")
	if ticketString == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "ticket_string is required in the URL path")
	}

	resp, err := h.svc.GetPasaje(c.Request().Context(), models.GetPasajeRequest{TicketString: ticketString})
	if err != nil {
		return err
	}

	if len(resp.Body) == 0 {
		return c.NoContent(resp.StatusCode)
	}

	if json.Valid(resp.Body) {
		return c.JSONBlob(resp.StatusCode, resp.Body)
	}

	return c.JSON(resp.StatusCode, map[string]string{"data": string(resp.Body)})
}
