package consumers

import (
	"net/http"
	"vm-server/models"
	services "vm-server/services/scan"

	"github.com/labstack/echo/v4"
)

type Handler struct {
	Service *services.Service
}

func NewHandler(s *services.Service) *Handler { return &Handler{Service: s} }

// ListConsumersHandler returns consumers, optionally filtered by filePath
func (h *Handler) ListConsumersHandler(c echo.Context) error {
	fp := c.QueryParam("filePath")
	out, err := h.Service.ListConsumers(&models.ConsumerFilter{FilePath: fp})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to list consumers"})
	}
	return c.JSON(http.StatusOK, out)
}
