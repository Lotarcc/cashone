package middleware

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"cashone/domain/entity"
	"cashone/domain/service"
)

const (
	authorizationHeader = "Authorization"
	bearerPrefix        = "Bearer "
	userContextKey      = "user"
)

// AuthMiddleware handles authentication for HTTP requests
type AuthMiddleware struct {
	authService service.AuthService
	log         *zap.SugaredLogger
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(authService service.AuthService, log *zap.SugaredLogger) *AuthMiddleware {
	return &AuthMiddleware{
		authService: authService,
		log:         log,
	}
}

// Authenticate is a middleware that validates JWT tokens and sets user claims in context
func (m *AuthMiddleware) Authenticate(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		auth := c.Request().Header.Get(authorizationHeader)
		if auth == "" {
			return echo.NewHTTPError(http.StatusUnauthorized, "Missing authorization header")
		}

		if !strings.HasPrefix(auth, bearerPrefix) {
			return echo.NewHTTPError(http.StatusUnauthorized, "Invalid authorization header format")
		}

		token := strings.TrimPrefix(auth, bearerPrefix)
		claims, err := m.authService.ValidateToken(c.Request().Context(), token)
		if err != nil {
			m.log.Errorw("Failed to validate token",
				"error", err,
			)
			return echo.NewHTTPError(http.StatusUnauthorized, "Invalid token")
		}

		// Store claims in context
		c.Set(userContextKey, claims)
		return next(c)
	}
}

// GetUserFromContext retrieves the user claims from the context
func GetUserFromContext(c echo.Context) *entity.Claims {
	user, ok := c.Get(userContextKey).(*entity.Claims)
	if !ok {
		return nil
	}
	return user
}

// GetUserIDFromContext retrieves the user ID from the context
func GetUserIDFromContext(c echo.Context) string {
	claims := GetUserFromContext(c)
	if claims == nil {
		return ""
	}
	return claims.UserID.String()
}
