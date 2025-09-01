package consumer

import (
	"net/http"
	"time"
	"vm-server/jobs"
	"vm-server/models"
	services "vm-server/services/consumer"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// Handler holds the service and scanner for the scan handlers.
type Handler struct {
	Service *services.Service
	Scanner *jobs.ConsumerScanner
}

// NewHandler creates a new scan handler.
func NewHandler(s *services.Service, sc *jobs.ConsumerScanner) *Handler {
	return &Handler{Service: s, Scanner: sc}
}

// ScanHandler handles scan related requests.
func (h *Handler) ConsumerHandler(c echo.Context) error {
	var req models.ConsumerRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request payload"})
	}

	jobID := uuid.New().String()
	newJob := &models.ConsumerJob{
		JobID:     jobID,
		Status:    "pending",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := h.Service.CreateConsumerJob(newJob); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create consumer job"})
	}

	// Run the scan in the background
	// TODO - Make this a channel so we can perform operations on failed scans as well.
			go h.Scanner.PerformScan(newJob, &req)

	return c.JSON(http.StatusAccepted, map[string]string{"jobId": jobID})
}

// ScanByIDHandler handles retrieving a scan by its ID.
func (h *Handler) ConsumerByIDHandler(c echo.Context) error {
	id := c.Param("id")
	consumerJob, err := h.Service.GetConsumerJob(id)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Consumer job not found"})
	}

	return c.JSON(http.StatusOK, consumerJob)
}
