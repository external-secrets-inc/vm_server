package secrets

import (
	"net/http"

	"vm-server/services/scan"
	"vm-server/services/secrets"

	"github.com/labstack/echo/v4"
)

// Handler holds the service and scanner for the scan handlers.
type Handler struct {
	Service *secrets.Service
	Scanner *scan.Service
}

func NewHandler(s *secrets.Service, sc *scan.Service) *Handler {
	return &Handler{Service: s, Scanner: sc}
}

// CreateSecretVersionHandler handles create secret version related requests
func (h *Handler) CreateSecretVersionHandler(c echo.Context) error {
	id := c.Param("id")
	req := updateVersionRequest{}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request payload"})
	}
	entry, err := h.Scanner.GetScanEntryByFingerprint(id)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Scan entry not found"})
	}
	err = h.Service.UpdateVersion([]byte(req.Value), entry)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update version"})
	}
	return c.JSON(http.StatusCreated, map[string]string{"message": "Version created"})
}
