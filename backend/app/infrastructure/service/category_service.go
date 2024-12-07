package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"cashone/domain/entity"
	"cashone/domain/repository"
	"cashone/domain/service"
)

var (
	ErrCategoryNotFound    = errors.New("category not found")
	ErrInvalidCategoryData = errors.New("invalid category data")
	ErrCategoryHasChildren = errors.New("cannot delete category with children")
	ErrCircularReference   = errors.New("circular reference in category hierarchy")
)

type categoryService struct {
	categoryRepo repository.CategoryRepository
	userRepo     repository.UserRepository
	log          *zap.SugaredLogger
}

func NewCategoryService(
	categoryRepo repository.CategoryRepository,
	userRepo repository.UserRepository,
	log *zap.SugaredLogger,
) service.CategoryService {
	return &categoryService{
		categoryRepo: categoryRepo,
		userRepo:     userRepo,
		log:          log,
	}
}

func (s *categoryService) CreateCategory(ctx context.Context, userID uuid.UUID, category *entity.Category) error {
	// Verify user exists
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		s.log.Errorw("Failed to get user for category creation", "error", err, "user_id", userID)
		return fmt.Errorf("failed to verify user: %w", err)
	}
	if user == nil {
		return ErrUserNotFound
	}

	// Validate category data
	if err := s.validateCategoryData(ctx, category); err != nil {
		return err
	}

	// Set category properties
	category.UserID = userID

	// If parent category is specified, verify it exists and belongs to the user
	if category.ParentID != nil {
		parent, err := s.categoryRepo.GetByID(ctx, *category.ParentID)
		if err != nil {
			s.log.Errorw("Failed to get parent category",
				"error", err,
				"parent_id", category.ParentID,
			)
			return fmt.Errorf("failed to verify parent category: %w", err)
		}
		if parent == nil {
			return ErrCategoryNotFound
		}
		if parent.UserID != userID {
			return ErrUnauthorized
		}
		if parent.Type != category.Type {
			return fmt.Errorf("%w: parent and child category types must match", ErrInvalidCategoryData)
		}
	}

	// Create the category
	if err := s.categoryRepo.Create(ctx, category); err != nil {
		s.log.Errorw("Failed to create category",
			"error", err,
			"user_id", userID,
			"name", category.Name,
		)
		return fmt.Errorf("failed to create category: %w", err)
	}

	return nil
}

func (s *categoryService) GetUserCategories(ctx context.Context, userID uuid.UUID) ([]entity.Category, error) {
	// Verify user exists
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		s.log.Errorw("Failed to get user for categories retrieval", "error", err, "user_id", userID)
		return nil, fmt.Errorf("failed to verify user: %w", err)
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	// Get all categories for the user
	categories, err := s.categoryRepo.GetByUserID(ctx, userID)
	if err != nil {
		s.log.Errorw("Failed to get user categories", "error", err, "user_id", userID)
		return nil, fmt.Errorf("failed to get categories: %w", err)
	}

	return categories, nil
}

func (s *categoryService) UpdateCategory(ctx context.Context, userID uuid.UUID, category *entity.Category) error {
	// Get existing category
	existingCategory, err := s.categoryRepo.GetByID(ctx, category.ID)
	if err != nil {
		s.log.Errorw("Failed to get category for update", "error", err, "category_id", category.ID)
		return fmt.Errorf("failed to get category: %w", err)
	}
	if existingCategory == nil {
		return ErrCategoryNotFound
	}

	// Verify category ownership
	if existingCategory.UserID != userID {
		return ErrUnauthorized
	}

	// Validate category data
	if err := s.validateCategoryData(ctx, category); err != nil {
		return err
	}

	// If parent category is changing, verify the new parent
	if category.ParentID != nil && (existingCategory.ParentID == nil || *category.ParentID != *existingCategory.ParentID) {
		parent, err := s.categoryRepo.GetByID(ctx, *category.ParentID)
		if err != nil {
			s.log.Errorw("Failed to get parent category",
				"error", err,
				"parent_id", category.ParentID,
			)
			return fmt.Errorf("failed to verify parent category: %w", err)
		}
		if parent == nil {
			return ErrCategoryNotFound
		}
		if parent.UserID != userID {
			return ErrUnauthorized
		}
		if parent.Type != category.Type {
			return fmt.Errorf("%w: parent and child category types must match", ErrInvalidCategoryData)
		}
	}

	// Preserve unchangeable fields
	category.UserID = existingCategory.UserID

	// Update the category
	if err := s.categoryRepo.Update(ctx, category); err != nil {
		s.log.Errorw("Failed to update category",
			"error", err,
			"category_id", category.ID,
			"user_id", userID,
		)
		return fmt.Errorf("failed to update category: %w", err)
	}

	return nil
}

func (s *categoryService) DeleteCategory(ctx context.Context, userID, categoryID uuid.UUID) error {
	// Get existing category
	category, err := s.categoryRepo.GetByID(ctx, categoryID)
	if err != nil {
		s.log.Errorw("Failed to get category for deletion", "error", err, "category_id", categoryID)
		return fmt.Errorf("failed to get category: %w", err)
	}
	if category == nil {
		return ErrCategoryNotFound
	}

	// Verify category ownership
	if category.UserID != userID {
		return ErrUnauthorized
	}

	// Delete the category
	if err := s.categoryRepo.Delete(ctx, categoryID); err != nil {
		s.log.Errorw("Failed to delete category",
			"error", err,
			"category_id", categoryID,
			"user_id", userID,
		)
		return fmt.Errorf("failed to delete category: %w", err)
	}

	return nil
}

func (s *categoryService) validateCategoryData(ctx context.Context, category *entity.Category) error {
	if category.Name == "" {
		return fmt.Errorf("%w: name is required", ErrInvalidCategoryData)
	}

	if category.Type != "expense" && category.Type != "income" && category.Type != "transfer" {
		return fmt.Errorf("%w: invalid category type", ErrInvalidCategoryData)
	}

	return nil
}
