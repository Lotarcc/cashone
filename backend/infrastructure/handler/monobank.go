package handler

import (
	"io"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"cashone/domain/errors"
	"cashone/domain/service"
	"cashone/infrastructure/middleware"
)

// MonobankHandler handles HTTP requests for Monobank integration endpoints
type MonobankHandler struct {
	log             *zap.SugaredLogger
	monobankService service.MonobankService
}

// NewMonobankHandler creates a new monobank handler and registers routes
func NewMonobankHandler(
	e *echo.Echo,
	log *zap.SugaredLogger,
	monobankService service.MonobankService,
	authMiddleware *middleware.AuthMiddleware,
) *MonobankHandler {
	handler := &MonobankHandler{
		log:             log,
		monobankService: monobankService,
	}

	monobank := e.Group("/api/v1/monobank")
	monobank.Use(authMiddleware.Authenticate)
	monobank.POST("/connect", handler.Connect)
	monobank.POST("/disconnect", handler.Disconnect)
	monobank.POST("/sync", handler.Sync)
	monobank.GET("/status", handler.Status)
	monobank.POST("/webhook", handler.Webhook)

	return handler
}

// Connect godoc
// @Summary Connect Monobank account
// @Description Connect user's Monobank account using personal token
// @Tags monobank
// @Accept json
// @Produce json
// @Param token body connectRequest true "Monobank personal token"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 429 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/monobank/connect [post]
// @Security Bearer
func (h *MonobankHandler) Connect(c echo.Context) error {
	var req connectRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	userIDStr := middleware.GetUserIDFromContext(c)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid user ID")
	}

	if err := h.monobankService.Connect(c.Request().Context(), userID, req.Token); err != nil {
		switch err {
		case errors.ErrMonobankTokenInvalid:
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid Monobank token")
		case errors.ErrMonobankRateLimit:
			return echo.NewHTTPError(http.StatusTooManyRequests, "Rate limit exceeded")
		case errors.ErrMonobankAlreadyConnected:
			return echo.NewHTTPError(http.StatusBadRequest, "Monobank already connected")
		default:
			h.log.Errorw("Failed to connect Monobank account",
				"error", err,
				"user_id", userID,
			)
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to connect Monobank account")
		}
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Successfully connected Monobank account",
	})
}

// Disconnect godoc
// @Summary Disconnect Monobank account
// @Description Disconnect user's Monobank account
// @Tags monobank
// @Accept json
// @Produce json
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/monobank/disconnect [post]
// @Security Bearer
func (h *MonobankHandler) Disconnect(c echo.Context) error {
	userIDStr := middleware.GetUserIDFromContext(c)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid user ID")
	}

	if err := h.monobankService.Disconnect(c.Request().Context(), userID); err != nil {
		switch err {
		case errors.ErrMonobankIntegrationNotFound:
			return echo.NewHTTPError(http.StatusNotFound, "Monobank integration not found")
		default:
			h.log.Errorw("Failed to disconnect Monobank account",
				"error", err,
				"user_id", userID,
			)
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to disconnect Monobank account")
		}
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Successfully disconnected Monobank account",
	})
}

// Sync godoc
// @Summary Sync Monobank data
// @Description Manually trigger synchronization of Monobank data
// @Tags monobank
// @Accept json
// @Produce json
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 429 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/monobank/sync [post]
// @Security Bearer
func (h *MonobankHandler) Sync(c echo.Context) error {
	userIDStr := middleware.GetUserIDFromContext(c)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid user ID")
	}

	if err := h.monobankService.SyncUserData(c.Request().Context(), userID); err != nil {
		switch err {
		case errors.ErrMonobankIntegrationNotFound:
			return echo.NewHTTPError(http.StatusNotFound, "Monobank integration not found")
		case errors.ErrMonobankRateLimit:
			return echo.NewHTTPError(http.StatusTooManyRequests, "Rate limit exceeded")
		default:
			h.log.Errorw("Failed to sync Monobank data",
				"error", err,
				"user_id", userID,
			)
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to sync Monobank data")
		}
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Successfully synced Monobank data",
	})
}

// Status godoc
// @Summary Get Monobank integration status
// @Description Get current status of user's Monobank integration
// @Tags monobank
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=entity.MonobankIntegration}
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/monobank/status [get]
// @Security Bearer
func (h *MonobankHandler) Status(c echo.Context) error {
	userIDStr := middleware.GetUserIDFromContext(c)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid user ID")
	}

	integration, err := h.monobankService.GetStatus(c.Request().Context(), userID)
	if err != nil {
		switch err {
		case errors.ErrMonobankIntegrationNotFound:
			return echo.NewHTTPError(http.StatusNotFound, "Monobank integration not found")
		default:
			h.log.Errorw("Failed to get Monobank integration status",
				"error", err,
				"user_id", userID,
			)
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get Monobank integration status")
		}
	}

	return c.JSON(http.StatusOK, integration)
}

// Webhook godoc
// @Summary Handle Monobank webhook
// @Description Handle webhook notifications from Monobank
// @Tags monobank
// @Accept json
// @Produce json
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/monobank/webhook [post]
func (h *MonobankHandler) Webhook(c echo.Context) error {
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		h.log.Errorw("Failed to read webhook body",
			"error", err,
		)
		return echo.NewHTTPError(http.StatusBadRequest, "Failed to read request body")
	}

	if err := h.monobankService.HandleWebhook(c.Request().Context(), body); err != nil {
		h.log.Errorw("Failed to handle webhook",
			"error", err,
		)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to handle webhook")
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Successfully handled webhook",
	})
}

// connectRequest represents the request body for connecting a Monobank account
type connectRequest struct {
	Token string `json:"token" validate:"required"`
}
