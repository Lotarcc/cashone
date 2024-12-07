package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"cashone/infrastructure/handler/response"
	"cashone/pkg/version"
)

func TestHealthCheck(t *testing.T) {
	// Setup
	logger := zap.NewExample().Sugar()
	e := echo.New()
	NewHealthHandler(e, logger)

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	// Perform request
	e.ServeHTTP(rec, req)

	// Assert status code
	assert.Equal(t, http.StatusOK, rec.Code)

	// Parse response
	var resp response.Response
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)
	require.True(t, resp.Success)
	require.NotNil(t, resp.Data)

	// Parse health data
	var healthData struct {
		Status  string `json:"status"`
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
	}

	healthDataBytes, err := json.Marshal(resp.Data)
	require.NoError(t, err)
	err = json.Unmarshal(healthDataBytes, &healthData)
	require.NoError(t, err)

	// Assert health check data
	assert.Equal(t, "ok", healthData.Status)

	// Assert version information
	versionInfo := version.GetInfo()
	assert.Equal(t, versionInfo.Version, healthData.Version.Current)
	assert.Equal(t, versionInfo.GitCommit, healthData.Version.Commit)
	assert.Equal(t, versionInfo.BuildTime, healthData.Version.BuildTime)
	assert.Equal(t, versionInfo.GoVersion, healthData.Version.GoVersion)
	assert.Equal(t, versionInfo.Platform, healthData.Version.Platform)

	// Assert runtime data
	assert.Greater(t, healthData.Runtime.NumGoroutine, 0)
	assert.NotEmpty(t, healthData.Runtime.GOOS)
	assert.NotEmpty(t, healthData.Runtime.GOARCH)
	assert.Greater(t, healthData.Runtime.NumCPU, 0)
	assert.Greater(t, healthData.Runtime.MemStats.Alloc, uint64(0))
	assert.Greater(t, healthData.Runtime.MemStats.TotalAlloc, uint64(0))
	assert.Greater(t, healthData.Runtime.MemStats.Sys, uint64(0))
}

func TestHealthCheckContentType(t *testing.T) {
	// Setup
	logger := zap.NewExample().Sugar()
	e := echo.New()
	NewHealthHandler(e, logger)

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	// Perform request
	e.ServeHTTP(rec, req)

	// Assert content type
	assert.Equal(t, "application/json; charset=UTF-8", rec.Header().Get("Content-Type"))
}
