package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"cashone/domain/entity"
	"cashone/domain/errors"
	"cashone/domain/service"
	"cashone/infrastructure/middleware"
)

// TransactionHandler handles HTTP requests for transaction-related endpoints
type TransactionHandler struct {
	log                *zap.SugaredLogger
	transactionService service.TransactionService
}

// NewTransactionHandler creates a new transaction handler and registers routes
func NewTransactionHandler(
	e *echo.Echo,
	log *zap.SugaredLogger,
	transactionService service.TransactionService,
	authMiddleware *middleware.AuthMiddleware,
) *TransactionHandler {
	handler := &TransactionHandler{
		log:                log,
		transactionService: transactionService,
	}

	// All transaction routes require authentication
	transactions := e.Group("/api/v1/transactions", authMiddleware.Authenticate)
	transactions.POST("", handler.Create)
	transactions.GET("", handler.List)
	transactions.GET("/:id", handler.Get)
	transactions.PUT("/:id", handler.Update)
	transactions.DELETE("/:id", handler.Delete)
	transactions.GET("/search", handler.Search)

	return handler
}

// Create godoc
// @Summary Create a new transaction
// @Description Create a new transaction for the authenticated user
// @Tags transactions
// @Accept json
// @Produce json
// @Param transaction body createTransactionRequest true "Transaction details"
// @Success 200 {object} entity.Transaction
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/transactions [post]
// @Security Bearer
func (h *TransactionHandler) Create(c echo.Context) error {
	var req createTransactionRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	userIDStr := middleware.GetUserIDFromContext(c)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid user ID")
	}

	// Create transaction entity
	transaction := &entity.Transaction{
		UserID:          userID,
		CardID:          req.CardID,
		CategoryID:      req.CategoryID,
		Amount:          req.Amount,
		Type:            req.Type,
		Description:     req.Description,
		TransactionDate: req.TransactionDate,
		Comment:         req.Comment,
	}

	if err := h.transactionService.Create(c.Request().Context(), transaction); err != nil {
		h.log.Errorw("Failed to create transaction",
			"error", err,
			"user_id", userID,
		)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create transaction")
	}

	return c.JSON(http.StatusOK, transaction)
}

// List godoc
// @Summary List transactions
// @Description Get paginated list of transactions for the authenticated user
// @Tags transactions
// @Accept json
// @Produce json
// @Param page query int false "Page number (default: 1)"
// @Param limit query int false "Items per page (default: 20)"
// @Success 200 {array} entity.Transaction
// @Failure 401 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/transactions [get]
// @Security Bearer
func (h *TransactionHandler) List(c echo.Context) error {
	userIDStr := middleware.GetUserIDFromContext(c)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid user ID")
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.QueryParam("page"))
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}

	offset := (page - 1) * limit

	transactions, err := h.transactionService.GetByUserID(c.Request().Context(), userID, limit, offset)
	if err != nil {
		h.log.Errorw("Failed to get transactions",
			"error", err,
			"user_id", userID,
		)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get transactions")
	}

	return c.JSON(http.StatusOK, transactions)
}

// Get godoc
// @Summary Get transaction by ID
// @Description Get a specific transaction by its ID
// @Tags transactions
// @Accept json
// @Produce json
// @Param id path string true "Transaction ID"
// @Success 200 {object} entity.Transaction
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/transactions/{id} [get]
// @Security Bearer
func (h *TransactionHandler) Get(c echo.Context) error {
	userIDStr := middleware.GetUserIDFromContext(c)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid user ID")
	}

	transactionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid transaction ID")
	}

	transaction, err := h.transactionService.GetByID(c.Request().Context(), transactionID)
	if err != nil {
		switch err {
		case errors.ErrTransactionNotFound:
			return echo.NewHTTPError(http.StatusNotFound, "Transaction not found")
		default:
			h.log.Errorw("Failed to get transaction",
				"error", err,
				"transaction_id", transactionID,
				"user_id", userID,
			)
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get transaction")
		}
	}

	// Verify transaction belongs to user
	if transaction.UserID != userID {
		return echo.NewHTTPError(http.StatusNotFound, "Transaction not found")
	}

	return c.JSON(http.StatusOK, transaction)
}

// Update godoc
// @Summary Update transaction
// @Description Update an existing transaction
// @Tags transactions
// @Accept json
// @Produce json
// @Param id path string true "Transaction ID"
// @Param transaction body updateTransactionRequest true "Transaction details"
// @Success 200 {object} entity.Transaction
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/transactions/{id} [put]
// @Security Bearer
func (h *TransactionHandler) Update(c echo.Context) error {
	userIDStr := middleware.GetUserIDFromContext(c)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid user ID")
	}

	transactionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid transaction ID")
	}

	var req updateTransactionRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	// Get existing transaction
	transaction, err := h.transactionService.GetByID(c.Request().Context(), transactionID)
	if err != nil {
		switch err {
		case errors.ErrTransactionNotFound:
			return echo.NewHTTPError(http.StatusNotFound, "Transaction not found")
		default:
			h.log.Errorw("Failed to get transaction",
				"error", err,
				"transaction_id", transactionID,
				"user_id", userID,
			)
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get transaction")
		}
	}

	// Verify transaction belongs to user
	if transaction.UserID != userID {
		return echo.NewHTTPError(http.StatusNotFound, "Transaction not found")
	}

	// Update fields
	transaction.CategoryID = req.CategoryID
	transaction.Amount = req.Amount
	transaction.Type = req.Type
	transaction.Description = req.Description
	transaction.TransactionDate = req.TransactionDate
	transaction.Comment = req.Comment

	if err := h.transactionService.Update(c.Request().Context(), transaction); err != nil {
		h.log.Errorw("Failed to update transaction",
			"error", err,
			"transaction_id", transactionID,
			"user_id", userID,
		)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update transaction")
	}

	return c.JSON(http.StatusOK, transaction)
}

// Delete godoc
// @Summary Delete transaction
// @Description Delete an existing transaction
// @Tags transactions
// @Accept json
// @Produce json
// @Param id path string true "Transaction ID"
// @Success 200 {object} messageResponse
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/transactions/{id} [delete]
// @Security Bearer
func (h *TransactionHandler) Delete(c echo.Context) error {
	userIDStr := middleware.GetUserIDFromContext(c)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid user ID")
	}

	transactionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid transaction ID")
	}

	// Get existing transaction
	transaction, err := h.transactionService.GetByID(c.Request().Context(), transactionID)
	if err != nil {
		switch err {
		case errors.ErrTransactionNotFound:
			return echo.NewHTTPError(http.StatusNotFound, "Transaction not found")
		default:
			h.log.Errorw("Failed to get transaction",
				"error", err,
				"transaction_id", transactionID,
				"user_id", userID,
			)
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get transaction")
		}
	}

	// Verify transaction belongs to user
	if transaction.UserID != userID {
		return echo.NewHTTPError(http.StatusNotFound, "Transaction not found")
	}

	if err := h.transactionService.Delete(c.Request().Context(), transactionID); err != nil {
		h.log.Errorw("Failed to delete transaction",
			"error", err,
			"transaction_id", transactionID,
			"user_id", userID,
		)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to delete transaction")
	}

	return c.JSON(http.StatusOK, messageResponse{
		Message: "Transaction successfully deleted",
	})
}

// Search godoc
// @Summary Search transactions
// @Description Search transactions with filters
// @Tags transactions
// @Accept json
// @Produce json
// @Param q query string false "Search query in description"
// @Param type query string false "Transaction type (expense/income/transfer)"
// @Param category_id query string false "Category ID"
// @Param card_id query string false "Card ID"
// @Param from query string false "Start date (YYYY-MM-DD)"
// @Param to query string false "End date (YYYY-MM-DD)"
// @Param min_amount query number false "Minimum amount"
// @Param max_amount query number false "Maximum amount"
// @Param page query int false "Page number (default: 1)"
// @Param limit query int false "Items per page (default: 20)"
// @Success 200 {array} entity.Transaction
// @Failure 401 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/transactions/search [get]
// @Security Bearer
func (h *TransactionHandler) Search(c echo.Context) error {
	userIDStr := middleware.GetUserIDFromContext(c)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid user ID")
	}

	// Parse search filters
	filters := searchFilters{
		Query:      c.QueryParam("q"),
		Type:       c.QueryParam("type"),
		CategoryID: parseUUID(c.QueryParam("category_id")),
		CardID:     parseUUID(c.QueryParam("card_id")),
		FromDate:   parseDate(c.QueryParam("from")),
		ToDate:     parseDate(c.QueryParam("to")),
		MinAmount:  parseInt64(c.QueryParam("min_amount")),
		MaxAmount:  parseInt64(c.QueryParam("max_amount")),
		Page:       parseInt(c.QueryParam("page"), 1),
		Limit:      parseInt(c.QueryParam("limit"), 20),
	}

	// Validate filters
	if err := validateSearchFilters(&filters); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// Calculate offset
	offset := (filters.Page - 1) * filters.Limit

	// Search transactions
	transactions, err := h.transactionService.Search(c.Request().Context(), userID, filters.toSearchParams(), filters.Limit, offset)
	if err != nil {
		h.log.Errorw("Failed to search transactions",
			"error", err,
			"user_id", userID,
			"filters", filters,
		)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to search transactions")
	}

	return c.JSON(http.StatusOK, transactions)
}

func validateSearchFilters(filters *searchFilters) error {
	// Validate transaction type if provided
	if filters.Type != "" && filters.Type != "expense" && filters.Type != "income" && filters.Type != "transfer" {
		return errors.ErrInvalidFieldValue
	}

	// Validate date range
	if filters.FromDate != nil && filters.ToDate != nil && filters.FromDate.After(*filters.ToDate) {
		return errors.ErrInvalidFieldValue
	}

	// Validate amount range
	if filters.MinAmount != nil && filters.MaxAmount != nil && *filters.MinAmount > *filters.MaxAmount {
		return errors.ErrInvalidFieldValue
	}

	// Validate pagination
	if filters.Page < 1 {
		filters.Page = 1
	}
	if filters.Limit < 1 {
		filters.Limit = 20
	}
	if filters.Limit > 100 {
		filters.Limit = 100 // Maximum limit to prevent large queries
	}

	return nil
}

// Helper functions for parsing query parameters
func parseUUID(s string) *uuid.UUID {
	if s == "" {
		return nil
	}
	if id, err := uuid.Parse(s); err == nil {
		return &id
	}
	return nil
}

func parseDate(s string) *time.Time {
	if s == "" {
		return nil
	}
	if t, err := time.Parse("2006-01-02", s); err == nil {
		return &t
	}
	return nil
}

func parseInt64(s string) *int64 {
	if s == "" {
		return nil
	}
	if n, err := strconv.ParseInt(s, 10, 64); err == nil {
		return &n
	}
	return nil
}

func parseInt(s string, defaultValue int) int {
	if s == "" {
		return defaultValue
	}
	if n, err := strconv.Atoi(s); err == nil {
		return n
	}
	return defaultValue
}

// searchFilters represents the search parameters for filtering transactions
type searchFilters struct {
	Query      string
	Type       string
	CategoryID *uuid.UUID
	CardID     *uuid.UUID
	FromDate   *time.Time
	ToDate     *time.Time
	MinAmount  *int64
	MaxAmount  *int64
	Page       int
	Limit      int
}

func (f *searchFilters) toSearchParams() entity.TransactionSearchParams {
	return entity.TransactionSearchParams{
		Query:      f.Query,
		Type:       f.Type,
		CategoryID: f.CategoryID,
		CardID:     f.CardID,
		FromDate:   f.FromDate,
		ToDate:     f.ToDate,
		MinAmount:  f.MinAmount,
		MaxAmount:  f.MaxAmount,
	}
}

// createTransactionRequest represents the request body for creating a new transaction
type createTransactionRequest struct {
	CardID          uuid.UUID  `json:"card_id" validate:"required"`
	CategoryID      *uuid.UUID `json:"category_id"`
	Amount          int64      `json:"amount" validate:"required"`
	Type            string     `json:"type" validate:"required,oneof=expense income transfer"`
	Description     string     `json:"description" validate:"required"`
	TransactionDate time.Time  `json:"transaction_date" validate:"required"`
	Comment         string     `json:"comment"`
}

// updateTransactionRequest represents the request body for updating an existing transaction
type updateTransactionRequest struct {
	CategoryID      *uuid.UUID `json:"category_id"`
	Amount          int64      `json:"amount" validate:"required"`
	Type            string     `json:"type" validate:"required,oneof=expense income transfer"`
	Description     string     `json:"description" validate:"required"`
	TransactionDate time.Time  `json:"transaction_date" validate:"required"`
	Comment         string     `json:"comment"`
}
