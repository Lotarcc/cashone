package errors

import "errors"

// Common domain errors
var (
	// User errors
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrInvalidUserData   = errors.New("invalid user data")

	// Card errors
	ErrCardNotFound      = errors.New("card not found")
	ErrCardAlreadyExists = errors.New("card already exists")
	ErrInvalidCardData   = errors.New("invalid card data")

	// Transaction errors
	ErrTransactionNotFound    = errors.New("transaction not found")
	ErrInvalidTransactionData = errors.New("invalid transaction data")

	// Category errors
	ErrCategoryNotFound      = errors.New("category not found")
	ErrCategoryAlreadyExists = errors.New("category already exists")
	ErrInvalidCategoryData   = errors.New("invalid category data")

	// Monobank errors
	ErrMonobankIntegrationNotFound = errors.New("monobank integration not found")
	ErrMonobankAlreadyConnected    = errors.New("monobank already connected")
	ErrMonobankTokenInvalid        = errors.New("monobank token invalid")
	ErrMonobankAPIError            = errors.New("monobank API error")
	ErrMonobankRateLimit           = errors.New("monobank rate limit exceeded")

	// Authentication errors
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrTokenExpired       = errors.New("token expired")
	ErrInvalidToken       = errors.New("invalid token")
	ErrUnauthorized       = errors.New("unauthorized")

	// Validation errors
	ErrValidation        = errors.New("validation error")
	ErrMissingField      = errors.New("missing required field")
	ErrInvalidFieldValue = errors.New("invalid field value")

	// Database errors
	ErrDatabaseConnection = errors.New("database connection error")
	ErrDatabaseOperation  = errors.New("database operation error")

	// General errors
	ErrInternal         = errors.New("internal error")
	ErrNotImplemented   = errors.New("not implemented")
	ErrInvalidRequest   = errors.New("invalid request")
	ErrResourceNotFound = errors.New("resource not found")
)
