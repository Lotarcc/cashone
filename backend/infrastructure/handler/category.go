package handler

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"cashone/domain/entity"
	"cashone/domain/errors"
	"cashone/domain/service"
	"cashone/infrastructure/handler/response"
	"cashone/infrastructure/middleware"
)

// CategoryHandler handles HTTP requests for category-related endpoints
type CategoryHandler struct {
	log             *zap.SugaredLogger
	categoryService service.CategoryService
}

// NewCategoryHandler creates a new category handler and registers routes
func NewCategoryHandler(
	e *echo.Echo,
	log *zap.SugaredLogger,
	categoryService service.CategoryService,
	authMiddleware *middleware.AuthMiddleware,
) *CategoryHandler {
	handler := &CategoryHandler{
		log:             log,
		categoryService: categoryService,
	}

	// All category routes require authentication
	categories := e.Group("/api/v1/categories", authMiddleware.Authenticate)
	categories.POST("", handler.Create)
	categories.GET("", handler.List)
	categories.GET("/:id", handler.Get)
	categories.PUT("/:id", handler.Update)
	categories.DELETE("/:id", handler.Delete)
	categories.GET("/tree", handler.GetTree)
	categories.GET("/:id/children", handler.GetChildren)
	categories.PUT("/:id/move", handler.Move)
	categories.POST("/default", handler.CreateDefault)

	return handler
}

// Create godoc
// @Summary Create a new category
// @Description Create a new category for the authenticated user
// @Tags categories
// @Accept json
// @Produce json
// @Param category body createCategoryRequest true "Category details"
// @Success 200 {object} response.Response{data=entity.Category}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/categories [post]
// @Security Bearer
func (h *CategoryHandler) Create(c echo.Context) error {
	var req createCategoryRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, response.NewErrorResponse("INVALID_REQUEST", "Invalid request body", err.Error()))
	}

	userIDStr := middleware.GetUserIDFromContext(c)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, response.NewErrorResponse("UNAUTHORIZED", "Invalid user ID", err.Error()))
	}

	category := &entity.Category{
		Base: entity.Base{
			ID: uuid.New(),
		},
		Name:     req.Name,
		Type:     req.Type,
		ParentID: req.ParentID,
		UserID:   userID,
	}

	if err := h.categoryService.Create(c.Request().Context(), category); err != nil {
		switch err {
		case errors.ErrCategoryAlreadyExists:
			return c.JSON(http.StatusBadRequest, response.NewErrorResponse("CATEGORY_EXISTS", "Category already exists", ""))
		default:
			h.log.Errorw("Failed to create category",
				"error", err,
				"user_id", userID,
			)
			return c.JSON(http.StatusInternalServerError, response.NewErrorResponse("INTERNAL_ERROR", "Failed to create category", ""))
		}
	}

	return c.JSON(http.StatusOK, response.NewResponse("Category created successfully", category))
}

// List godoc
// @Summary List categories
// @Description Get list of categories for the authenticated user
// @Tags categories
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=[]entity.Category}
// @Failure 401 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/categories [get]
// @Security Bearer
func (h *CategoryHandler) List(c echo.Context) error {
	userIDStr := middleware.GetUserIDFromContext(c)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, response.NewErrorResponse("UNAUTHORIZED", "Invalid user ID", err.Error()))
	}

	categories, err := h.categoryService.GetByUserID(c.Request().Context(), userID)
	if err != nil {
		h.log.Errorw("Failed to get categories",
			"error", err,
			"user_id", userID,
		)
		return c.JSON(http.StatusInternalServerError, response.NewErrorResponse("INTERNAL_ERROR", "Failed to get categories", ""))
	}

	return c.JSON(http.StatusOK, response.NewResponse("Categories retrieved successfully", categories))
}

// Get godoc
// @Summary Get category by ID
// @Description Get a specific category by its ID
// @Tags categories
// @Accept json
// @Produce json
// @Param id path string true "Category ID"
// @Success 200 {object} response.Response{data=entity.Category}
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/categories/{id} [get]
// @Security Bearer
func (h *CategoryHandler) Get(c echo.Context) error {
	userIDStr := middleware.GetUserIDFromContext(c)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, response.NewErrorResponse("UNAUTHORIZED", "Invalid user ID", err.Error()))
	}

	categoryID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.NewErrorResponse("INVALID_ID", "Invalid category ID", err.Error()))
	}

	category, err := h.categoryService.GetByID(c.Request().Context(), categoryID)
	if err != nil {
		switch err {
		case errors.ErrCategoryNotFound:
			return c.JSON(http.StatusNotFound, response.NewErrorResponse("NOT_FOUND", "Category not found", ""))
		default:
			h.log.Errorw("Failed to get category",
				"error", err,
				"category_id", categoryID,
				"user_id", userID,
			)
			return c.JSON(http.StatusInternalServerError, response.NewErrorResponse("INTERNAL_ERROR", "Failed to get category", ""))
		}
	}

	// Verify category belongs to user
	if category.UserID != userID {
		return c.JSON(http.StatusNotFound, response.NewErrorResponse("NOT_FOUND", "Category not found", ""))
	}

	return c.JSON(http.StatusOK, response.NewResponse("Category retrieved successfully", category))
}

// Update godoc
// @Summary Update category
// @Description Update an existing category
// @Tags categories
// @Accept json
// @Produce json
// @Param id path string true "Category ID"
// @Param category body updateCategoryRequest true "Category details"
// @Success 200 {object} response.Response{data=entity.Category}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/categories/{id} [put]
// @Security Bearer
func (h *CategoryHandler) Update(c echo.Context) error {
	userIDStr := middleware.GetUserIDFromContext(c)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, response.NewErrorResponse("UNAUTHORIZED", "Invalid user ID", err.Error()))
	}

	categoryID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.NewErrorResponse("INVALID_ID", "Invalid category ID", err.Error()))
	}

	var req updateCategoryRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, response.NewErrorResponse("INVALID_REQUEST", "Invalid request body", err.Error()))
	}

	category := &entity.Category{
		Base: entity.Base{
			ID: categoryID,
		},
		Name:     req.Name,
		Type:     req.Type,
		ParentID: req.ParentID,
		UserID:   userID,
	}

	if err := h.categoryService.Update(c.Request().Context(), category); err != nil {
		switch err {
		case errors.ErrCategoryNotFound:
			return c.JSON(http.StatusNotFound, response.NewErrorResponse("NOT_FOUND", "Category not found", ""))
		case errors.ErrUnauthorized:
			return c.JSON(http.StatusNotFound, response.NewErrorResponse("NOT_FOUND", "Category not found", ""))
		default:
			h.log.Errorw("Failed to update category",
				"error", err,
				"category_id", categoryID,
				"user_id", userID,
			)
			return c.JSON(http.StatusInternalServerError, response.NewErrorResponse("INTERNAL_ERROR", "Failed to update category", ""))
		}
	}

	return c.JSON(http.StatusOK, response.NewResponse("Category updated successfully", category))
}

// Delete godoc
// @Summary Delete category
// @Description Delete an existing category
// @Tags categories
// @Accept json
// @Produce json
// @Param id path string true "Category ID"
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/categories/{id} [delete]
// @Security Bearer
func (h *CategoryHandler) Delete(c echo.Context) error {
	userIDStr := middleware.GetUserIDFromContext(c)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, response.NewErrorResponse("UNAUTHORIZED", "Invalid user ID", err.Error()))
	}

	categoryID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.NewErrorResponse("INVALID_ID", "Invalid category ID", err.Error()))
	}

	// Get category first to verify ownership
	category, err := h.categoryService.GetByID(c.Request().Context(), categoryID)
	if err != nil {
		switch err {
		case errors.ErrCategoryNotFound:
			return c.JSON(http.StatusNotFound, response.NewErrorResponse("NOT_FOUND", "Category not found", ""))
		default:
			h.log.Errorw("Failed to get category",
				"error", err,
				"category_id", categoryID,
				"user_id", userID,
			)
			return c.JSON(http.StatusInternalServerError, response.NewErrorResponse("INTERNAL_ERROR", "Failed to delete category", ""))
		}
	}

	// Verify category belongs to user
	if category.UserID != userID {
		return c.JSON(http.StatusNotFound, response.NewErrorResponse("NOT_FOUND", "Category not found", ""))
	}

	if err := h.categoryService.Delete(c.Request().Context(), categoryID); err != nil {
		h.log.Errorw("Failed to delete category",
			"error", err,
			"category_id", categoryID,
			"user_id", userID,
		)
		return c.JSON(http.StatusInternalServerError, response.NewErrorResponse("INTERNAL_ERROR", "Failed to delete category", ""))
	}

	return c.JSON(http.StatusOK, response.NewResponse("Category deleted successfully", nil))
}

// GetTree godoc
// @Summary Get category hierarchy
// @Description Get hierarchical tree of categories for the authenticated user
// @Tags categories
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=[]entity.CategoryTree}
// @Failure 401 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/categories/tree [get]
// @Security Bearer
func (h *CategoryHandler) GetTree(c echo.Context) error {
	userIDStr := middleware.GetUserIDFromContext(c)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, response.NewErrorResponse("UNAUTHORIZED", "Invalid user ID", err.Error()))
	}

	tree, err := h.categoryService.GetTree(c.Request().Context(), userID)
	if err != nil {
		h.log.Errorw("Failed to get category tree",
			"error", err,
			"user_id", userID,
		)
		return c.JSON(http.StatusInternalServerError, response.NewErrorResponse("INTERNAL_ERROR", "Failed to get category tree", ""))
	}

	return c.JSON(http.StatusOK, response.NewResponse("Category tree retrieved successfully", tree))
}

// GetChildren godoc
// @Summary Get category children
// @Description Get direct children of a specific category
// @Tags categories
// @Accept json
// @Produce json
// @Param id path string true "Category ID"
// @Success 200 {object} response.Response{data=[]entity.Category}
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/categories/{id}/children [get]
// @Security Bearer
func (h *CategoryHandler) GetChildren(c echo.Context) error {
	userIDStr := middleware.GetUserIDFromContext(c)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, response.NewErrorResponse("UNAUTHORIZED", "Invalid user ID", err.Error()))
	}

	categoryID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.NewErrorResponse("INVALID_ID", "Invalid category ID", err.Error()))
	}

	// Get category first to verify ownership
	category, err := h.categoryService.GetByID(c.Request().Context(), categoryID)
	if err != nil {
		switch err {
		case errors.ErrCategoryNotFound:
			return c.JSON(http.StatusNotFound, response.NewErrorResponse("NOT_FOUND", "Category not found", ""))
		default:
			h.log.Errorw("Failed to get category",
				"error", err,
				"category_id", categoryID,
				"user_id", userID,
			)
			return c.JSON(http.StatusInternalServerError, response.NewErrorResponse("INTERNAL_ERROR", "Failed to get category children", ""))
		}
	}

	// Verify category belongs to user
	if category.UserID != userID {
		return c.JSON(http.StatusNotFound, response.NewErrorResponse("NOT_FOUND", "Category not found", ""))
	}

	children, err := h.categoryService.GetChildren(c.Request().Context(), categoryID)
	if err != nil {
		h.log.Errorw("Failed to get category children",
			"error", err,
			"category_id", categoryID,
			"user_id", userID,
		)
		return c.JSON(http.StatusInternalServerError, response.NewErrorResponse("INTERNAL_ERROR", "Failed to get category children", ""))
	}

	return c.JSON(http.StatusOK, response.NewResponse("Category children retrieved successfully", children))
}

// Move godoc
// @Summary Move category
// @Description Move a category to a new parent
// @Tags categories
// @Accept json
// @Produce json
// @Param id path string true "Category ID"
// @Param request body moveCategoryRequest true "Move details"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/categories/{id}/move [put]
// @Security Bearer
func (h *CategoryHandler) Move(c echo.Context) error {
	userIDStr := middleware.GetUserIDFromContext(c)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, response.NewErrorResponse("UNAUTHORIZED", "Invalid user ID", err.Error()))
	}

	categoryID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.NewErrorResponse("INVALID_ID", "Invalid category ID", err.Error()))
	}

	var req moveCategoryRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, response.NewErrorResponse("INVALID_REQUEST", "Invalid request body", err.Error()))
	}

	// Get category first to verify ownership
	category, err := h.categoryService.GetByID(c.Request().Context(), categoryID)
	if err != nil {
		switch err {
		case errors.ErrCategoryNotFound:
			return c.JSON(http.StatusNotFound, response.NewErrorResponse("NOT_FOUND", "Category not found", ""))
		default:
			h.log.Errorw("Failed to get category",
				"error", err,
				"category_id", categoryID,
				"user_id", userID,
			)
			return c.JSON(http.StatusInternalServerError, response.NewErrorResponse("INTERNAL_ERROR", "Failed to move category", ""))
		}
	}

	// Verify category belongs to user
	if category.UserID != userID {
		return c.JSON(http.StatusNotFound, response.NewErrorResponse("NOT_FOUND", "Category not found", ""))
	}

	if err := h.categoryService.MoveCategory(c.Request().Context(), categoryID, req.ParentID); err != nil {
		switch err {
		case errors.ErrCategoryNotFound:
			return c.JSON(http.StatusNotFound, response.NewErrorResponse("NOT_FOUND", "Parent category not found", ""))
		case errors.ErrUnauthorized:
			return c.JSON(http.StatusBadRequest, response.NewErrorResponse("INVALID_OPERATION", "Cannot move category to another user's category", ""))
		case errors.ErrInvalidCategoryData:
			return c.JSON(http.StatusBadRequest, response.NewErrorResponse("INVALID_OPERATION", "Invalid move operation", ""))
		default:
			h.log.Errorw("Failed to move category",
				"error", err,
				"category_id", categoryID,
				"user_id", userID,
				"new_parent_id", req.ParentID,
			)
			return c.JSON(http.StatusInternalServerError, response.NewErrorResponse("INTERNAL_ERROR", "Failed to move category", ""))
		}
	}

	return c.JSON(http.StatusOK, response.NewResponse("Category moved successfully", nil))
}

// CreateDefault godoc
// @Summary Create default categories
// @Description Create default categories for the authenticated user
// @Tags categories
// @Accept json
// @Produce json
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/categories/default [post]
// @Security Bearer
func (h *CategoryHandler) CreateDefault(c echo.Context) error {
	userIDStr := middleware.GetUserIDFromContext(c)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, response.NewErrorResponse("UNAUTHORIZED", "Invalid user ID", err.Error()))
	}

	if err := h.categoryService.CreateDefaultCategories(c.Request().Context(), userID); err != nil {
		h.log.Errorw("Failed to create default categories",
			"error", err,
			"user_id", userID,
		)
		return c.JSON(http.StatusInternalServerError, response.NewErrorResponse("INTERNAL_ERROR", "Failed to create default categories", ""))
	}

	return c.JSON(http.StatusOK, response.NewResponse("Default categories created successfully", nil))
}

type createCategoryRequest struct {
	Name     string     `json:"name" validate:"required"`
	Type     string     `json:"type" validate:"required,oneof=expense income transfer"`
	ParentID *uuid.UUID `json:"parent_id"`
}

type updateCategoryRequest struct {
	Name     string     `json:"name" validate:"required"`
	Type     string     `json:"type" validate:"required,oneof=expense income transfer"`
	ParentID *uuid.UUID `json:"parent_id"`
}

type moveCategoryRequest struct {
	ParentID *uuid.UUID `json:"parent_id"`
}
