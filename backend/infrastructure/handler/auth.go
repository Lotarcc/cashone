package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"cashone/domain/entity"
	"cashone/domain/errors"
	"cashone/domain/service"
)

// AuthHandler handles HTTP requests for authentication-related endpoints
type AuthHandler struct {
	log         *zap.SugaredLogger
	authService service.AuthService
}

// NewAuthHandler creates a new auth handler and registers routes
func NewAuthHandler(
	e *echo.Echo,
	log *zap.SugaredLogger,
	authService service.AuthService,
) *AuthHandler {
	handler := &AuthHandler{
		log:         log,
		authService: authService,
	}

	auth := e.Group("/api/v1/auth")
	auth.POST("/register", handler.Register)
	auth.POST("/login", handler.Login)
	auth.POST("/refresh", handler.RefreshToken)
	auth.POST("/logout", handler.Logout)

	return handler
}

// Register godoc
// @Summary Register a new user
// @Description Register a new user with email and password
// @Tags auth
// @Accept json
// @Produce json
// @Param request body entity.RegisterRequest true "Registration details"
// @Success 200 {object} entity.RegisterResponse
// @Failure 400 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /api/v1/auth/register [post]
func (h *AuthHandler) Register(c echo.Context) error {
	var req entity.RegisterRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	// Validate request
	if req.Email == "" || req.Password == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Email and password are required")
	}

	// Register user
	resp, err := h.authService.Register(c.Request().Context(), &req)
	if err != nil {
		switch err {
		case errors.ErrUserAlreadyExists:
			return echo.NewHTTPError(http.StatusBadRequest, "User already exists")
		default:
			h.log.Errorw("Failed to register user",
				"error", err,
				"email", req.Email,
			)
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to register user")
		}
	}

	return c.JSON(http.StatusOK, resp)
}

// Login godoc
// @Summary Login user
// @Description Authenticate user with email and password
// @Tags auth
// @Accept json
// @Produce json
// @Param request body entity.LoginRequest true "Login credentials"
// @Success 200 {object} entity.LoginResponse
// @Failure 400 {object} echo.HTTPError
// @Failure 401 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /api/v1/auth/login [post]
func (h *AuthHandler) Login(c echo.Context) error {
	var req entity.LoginRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	// Validate request
	if req.Email == "" || req.Password == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Email and password are required")
	}

	// Login user
	resp, err := h.authService.Login(c.Request().Context(), &req)
	if err != nil {
		switch err {
		case errors.ErrInvalidCredentials:
			return echo.NewHTTPError(http.StatusUnauthorized, "Invalid email or password")
		default:
			h.log.Errorw("Failed to login user",
				"error", err,
				"email", req.Email,
			)
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to login user")
		}
	}

	return c.JSON(http.StatusOK, resp)
}

// RefreshToken godoc
// @Summary Refresh access token
// @Description Get new access token using refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body refreshTokenRequest true "Refresh token"
// @Success 200 {object} entity.AuthToken
// @Failure 400 {object} echo.HTTPError
// @Failure 401 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /api/v1/auth/refresh [post]
func (h *AuthHandler) RefreshToken(c echo.Context) error {
	var req refreshTokenRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	// Validate request
	if req.RefreshToken == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Refresh token is required")
	}

	// Refresh token
	token, err := h.authService.RefreshToken(c.Request().Context(), req.RefreshToken)
	if err != nil {
		switch err {
		case errors.ErrInvalidToken:
			return echo.NewHTTPError(http.StatusUnauthorized, "Invalid refresh token")
		case errors.ErrTokenExpired:
			return echo.NewHTTPError(http.StatusUnauthorized, "Refresh token expired")
		default:
			h.log.Errorw("Failed to refresh token",
				"error", err,
			)
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to refresh token")
		}
	}

	return c.JSON(http.StatusOK, token)
}

// Logout godoc
// @Summary Logout user
// @Description Revoke refresh token and logout user
// @Tags auth
// @Accept json
// @Produce json
// @Param request body logoutRequest true "Refresh token"
// @Success 200 {object} messageResponse
// @Failure 400 {object} echo.HTTPError
// @Failure 401 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /api/v1/auth/logout [post]
// @Security Bearer
func (h *AuthHandler) Logout(c echo.Context) error {
	var req logoutRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	// Get user ID from context
	claims := c.Get("user").(*entity.Claims)
	if claims == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
	}

	// Validate request
	if req.RefreshToken == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Refresh token is required")
	}

	// Logout user
	if err := h.authService.Logout(c.Request().Context(), claims.UserID, req.RefreshToken); err != nil {
		h.log.Errorw("Failed to logout user",
			"error", err,
			"user_id", claims.UserID,
		)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to logout user")
	}

	return c.JSON(http.StatusOK, messageResponse{
		Message: "Successfully logged out",
	})
}

type refreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type logoutRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type messageResponse struct {
	Message string `json:"message"`
}
