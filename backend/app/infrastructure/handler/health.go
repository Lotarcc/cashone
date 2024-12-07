package handler

import (
	"net/http"
	"runtime"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"cashone/infrastructure/handler/response"
	"cashone/pkg/version"
)

type HealthHandler struct {
	log *zap.SugaredLogger
}

// NewHealthHandler creates a new health check handler
func NewHealthHandler(e *echo.Echo, log *zap.SugaredLogger) {
	handler := &HealthHandler{
		log: log,
	}

	e.GET("/health", handler.Check)
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

	healthData := struct {
		response.HealthResponse
		Version struct {
			Current   string `json:"current"`
			Commit    string `json:"commit"`
			BuildTime string `json:"build_time"`
			GoVersion string `json:"go_version"`
			Platform  string `json:"platform"`
		} `json:"version"`
		Runtime struct {
			NumGoroutine int    `json:"num_goroutine"`
			GOOS         string `json:"goos"`
			GOARCH       string `json:"goarch"`
			NumCPU       int    `json:"num_cpu"`
			MemStats     struct {
				Alloc      uint64 `json:"alloc"`
				TotalAlloc uint64 `json:"total_alloc"`
				Sys        uint64 `json:"sys"`
				NumGC      uint32 `json:"num_gc"`
			} `json:"mem_stats"`
		} `json:"runtime"`
	}{}

	// Set basic health information
	healthData.Status = "ok"
	healthData.Version.Current = versionInfo.Version
	healthData.Version.Commit = versionInfo.GitCommit
	healthData.Version.BuildTime = versionInfo.BuildTime
	healthData.Version.GoVersion = versionInfo.GoVersion
	healthData.Version.Platform = versionInfo.Platform

	// Set runtime information
	healthData.Runtime.NumGoroutine = runtime.NumGoroutine()
	healthData.Runtime.GOOS = runtime.GOOS
	healthData.Runtime.GOARCH = runtime.GOARCH
	healthData.Runtime.NumCPU = runtime.NumCPU()
	healthData.Runtime.MemStats.Alloc = m.Alloc
	healthData.Runtime.MemStats.TotalAlloc = m.TotalAlloc
	healthData.Runtime.MemStats.Sys = m.Sys
	healthData.Runtime.MemStats.NumGC = m.NumGC

	h.log.Infow("Health check performed",
		"status", healthData.Status,
		"version", healthData.Version.Current,
		"goroutines", healthData.Runtime.NumGoroutine,
	)

	return c.JSON(http.StatusOK, response.NewResponse(healthData))
}
