package scan

import (
	"net/http"
	"time"
	"vm-server/jobs"
	"vm-server/models"
	services "vm-server/services/scan"
	"vm-server/watcher"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// Handler holds the service and scanner for the scan handlers.
type Handler struct {
	Service *services.Service
	Scanner *jobs.Scanner
	Watcher *watcher.Manager
}

// NewHandler creates a new scan handler.
func NewHandler(s *services.Service, sc *jobs.Scanner, w *watcher.Manager) *Handler {
	return &Handler{Service: s, Scanner: sc, Watcher: w}
}

// ScanHandler handles scan related requests.
func (h *Handler) ScanHandler(c echo.Context) error {
	var req models.ScanRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request payload"})
	}

	jobID := uuid.New().String()
	newJob := &models.ScanJob{
		JobID:     jobID,
		Status:    "pending",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := h.Service.CreateScanJob(newJob); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create scan job"})
	}

	// Activate watcher for requested paths (best-effort)
	if h.Watcher != nil {
		for _, p := range req.Paths {
			_ = h.Watcher.ActivatePath(p)
		}
	}

	// Run the scan in the background
	// TODO - Make this a channel so we can perform operations on failed scans as well.
	go h.Scanner.PerformScan(newJob, &req)

	return c.JSON(http.StatusAccepted, map[string]string{"jobId": jobID})
}

// ScanByIDHandler handles retrieving a scan by its ID.
func (h *Handler) ScanByIDHandler(c echo.Context) error {
	id := c.Param("id")
	scanJob, err := h.Service.GetScanJob(id)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Scan job not found"})
	}

	return c.JSON(http.StatusOK, scanJob)
}
