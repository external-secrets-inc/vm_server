package secrets

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// CreateSecretVersionHandler handles create secret version related requests
func CreateSecretVersionHandler(c echo.Context) error {
	return c.String(http.StatusOK, "ok")
}
