package handler

import (
	"net/http"
	"runtime"
	"time"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"cashone/domain/repository"
	"cashone/domain/service"
	"cashone/pkg/version"
)

// HealthHandler handles HTTP requests for health check endpoints
type HealthHandler struct {
	log            *zap.SugaredLogger
	repoFactory    repository.Factory
	serviceFactory service.Factory
}

// NewHealthHandler creates a new health check handler
func NewHealthHandler(
	e *echo.Echo,
	log *zap.SugaredLogger,
	repoFactory repository.Factory,
	serviceFactory service.Factory,
) *HealthHandler {
	handler := &HealthHandler{
		log:            log,
		repoFactory:    repoFactory,
		serviceFactory: serviceFactory,
	}

	e.GET("/health", handler.Check)
	return handler
}

// Check godoc
// @Summary Health check endpoint
// @Description Get server health status, version information, and basic metrics
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=response.HealthResponse}
// @Router /health [get]
func (h *HealthHandler) Check(c echo.Context) error {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	versionInfo := version.GetInfo()

	// Check database connection
	db := h.repoFactory.NewUserRepository()
	dbErr := db.Ping(c.Request().Context())

	healthData := struct {
		Status    string `json:"status"`
		Database  string `json:"database"`
		Version   string `json:"version"`
		Timestamp string `json:"timestamp"`
	}{
		Version:   versionInfo.Version,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	if dbErr != nil {
		healthData.Status = "degraded"
		healthData.Database = "error"
	} else {
		healthData.Status = "ok"
		healthData.Database = "ok"
	}

	h.log.Infow("Health check performed",
		"status", healthData.Status,
		"version", healthData.Version,
		"goroutines", runtime.NumGoroutine(),
		"database", healthData.Database,
	)

	return c.JSON(http.StatusOK, healthData)
}
