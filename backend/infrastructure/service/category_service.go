package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"cashone/domain/entity"
	"cashone/domain/errors"
	"cashone/domain/repository"
	"cashone/domain/service"
)

type categoryService struct {
	categoryRepo repository.CategoryRepository
	userRepo     repository.UserRepository
	log          *zap.SugaredLogger
}

// NewCategoryService creates a new category service
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

func (s *categoryService) Create(ctx context.Context, category *entity.Category) error {
	// Validate category data
	if err := s.validateCategory(category); err != nil {
		return fmt.Errorf("%w: %v", errors.ErrInvalidCategoryData, err)
	}

	// Check if user exists
	user, err := s.userRepo.GetByID(ctx, category.UserID)
	if err != nil {
		return fmt.Errorf("%w: %v", errors.ErrDatabaseOperation, err)
	}
	if user == nil {
		return errors.ErrUserNotFound
	}

	// Check if category with this name already exists for the user
	existingCategories, err := s.categoryRepo.GetByUserID(ctx, category.UserID)
	if err != nil {
		return fmt.Errorf("%w: %v", errors.ErrDatabaseOperation, err)
	}
	for _, existingCategory := range existingCategories {
		if existingCategory.Name == category.Name {
			return errors.ErrCategoryAlreadyExists
		}
	}

	// Generate UUID if not provided
	if category.ID == uuid.Nil {
		category.ID = uuid.New()
	}

	// Create category
	if err := s.categoryRepo.Create(ctx, category); err != nil {
		return fmt.Errorf("%w: %v", errors.ErrDatabaseOperation, err)
	}

	s.log.Infow("Category created successfully",
		"id", category.ID,
		"user_id", category.UserID,
		"name", category.Name,
	)
	return nil
}

func (s *categoryService) GetByID(ctx context.Context, id uuid.UUID) (*entity.Category, error) {
	category, err := s.categoryRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errors.ErrDatabaseOperation, err)
	}
	if category == nil {
		return nil, errors.ErrCategoryNotFound
	}
	return category, nil
}

func (s *categoryService) GetByUserID(ctx context.Context, userID uuid.UUID) ([]entity.Category, error) {
	// Check if user exists
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errors.ErrDatabaseOperation, err)
	}
	if user == nil {
		return nil, errors.ErrUserNotFound
	}

	// Get user's categories
	categories, err := s.categoryRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errors.ErrDatabaseOperation, err)
	}
	return categories, nil
}

func (s *categoryService) Update(ctx context.Context, category *entity.Category) error {
	// Validate category data
	if err := s.validateCategory(category); err != nil {
		return fmt.Errorf("%w: %v", errors.ErrInvalidCategoryData, err)
	}

	// Check if category exists
	existingCategory, err := s.categoryRepo.GetByID(ctx, category.ID)
	if err != nil {
		return fmt.Errorf("%w: %v", errors.ErrDatabaseOperation, err)
	}
	if existingCategory == nil {
		return errors.ErrCategoryNotFound
	}

	// Check if user owns the category
	if existingCategory.UserID != category.UserID {
		return errors.ErrUnauthorized
	}

	// Update category
	if err := s.categoryRepo.Update(ctx, category); err != nil {
		return fmt.Errorf("%w: %v", errors.ErrDatabaseOperation, err)
	}

	s.log.Infow("Category updated successfully",
		"id", category.ID,
		"user_id", category.UserID,
		"name", category.Name,
	)
	return nil
}

func (s *categoryService) Delete(ctx context.Context, id uuid.UUID) error {
	// Check if category exists
	existingCategory, err := s.categoryRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("%w: %v", errors.ErrDatabaseOperation, err)
	}
	if existingCategory == nil {
		return errors.ErrCategoryNotFound
	}

	// Delete category
	if err := s.categoryRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("%w: %v", errors.ErrDatabaseOperation, err)
	}

	s.log.Infow("Category deleted successfully", "id", id)
	return nil
}

func (s *categoryService) GetTree(ctx context.Context, userID uuid.UUID) ([]entity.CategoryTree, error) {
	// Get all categories for the user
	categories, err := s.categoryRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errors.ErrDatabaseOperation, err)
	}

	// Build category tree
	return s.buildCategoryTree(categories), nil
}

func (s *categoryService) GetChildren(ctx context.Context, categoryID uuid.UUID) ([]entity.Category, error) {
	// Get all categories for the user
	category, err := s.categoryRepo.GetByID(ctx, categoryID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errors.ErrDatabaseOperation, err)
	}
	if category == nil {
		return nil, errors.ErrCategoryNotFound
	}

	// Get all categories for the user
	allCategories, err := s.categoryRepo.GetByUserID(ctx, category.UserID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errors.ErrDatabaseOperation, err)
	}

	// Filter children
	var children []entity.Category
	for _, c := range allCategories {
		if c.ParentID != nil && *c.ParentID == categoryID {
			children = append(children, c)
		}
	}

	return children, nil
}

func (s *categoryService) MoveCategory(ctx context.Context, categoryID uuid.UUID, newParentID *uuid.UUID) error {
	// Get category
	category, err := s.categoryRepo.GetByID(ctx, categoryID)
	if err != nil {
		return fmt.Errorf("%w: %v", errors.ErrDatabaseOperation, err)
	}
	if category == nil {
		return errors.ErrCategoryNotFound
	}

	// If moving to a parent, verify parent exists and belongs to same user
	if newParentID != nil {
		parent, err := s.categoryRepo.GetByID(ctx, *newParentID)
		if err != nil {
			return fmt.Errorf("%w: %v", errors.ErrDatabaseOperation, err)
		}
		if parent == nil {
			return errors.ErrCategoryNotFound
		}
		if parent.UserID != category.UserID {
			return errors.ErrUnauthorized
		}

		// Prevent circular references
		if s.wouldCreateCircularReference(ctx, categoryID, *newParentID) {
			return errors.ErrInvalidCategoryData
		}
	}

	// Update category's parent
	category.ParentID = newParentID
	if err := s.categoryRepo.Update(ctx, category); err != nil {
		return fmt.Errorf("%w: %v", errors.ErrDatabaseOperation, err)
	}

	return nil
}

func (s *categoryService) CreateDefaultCategories(ctx context.Context, userID uuid.UUID) error {
	// Check if user exists
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("%w: %v", errors.ErrDatabaseOperation, err)
	}
	if user == nil {
		return errors.ErrUserNotFound
	}

	// Get default categories
	defaultCategories := s.GetDefaultCategories()

	// Create each category
	for _, category := range defaultCategories {
		category.UserID = userID
		category.ID = uuid.New()
		if err := s.categoryRepo.Create(ctx, &category); err != nil {
			return fmt.Errorf("%w: %v", errors.ErrDatabaseOperation, err)
		}
	}

	return nil
}

func (s *categoryService) GetDefaultCategories() []entity.Category {
	return []entity.Category{
		{Name: "Food & Dining", Type: "expense"},
		{Name: "Shopping", Type: "expense"},
		{Name: "Housing", Type: "expense"},
		{Name: "Transportation", Type: "expense"},
		{Name: "Vehicle", Type: "expense"},
		{Name: "Entertainment", Type: "expense"},
		{Name: "Healthcare", Type: "expense"},
		{Name: "Insurance", Type: "expense"},
		{Name: "Personal Care", Type: "expense"},
		{Name: "Education", Type: "expense"},
		{Name: "Gifts & Donations", Type: "expense"},
		{Name: "Investments", Type: "expense"},
		{Name: "Salary", Type: "income"},
		{Name: "Business", Type: "income"},
		{Name: "Gifts", Type: "income"},
		{Name: "Investments", Type: "income"},
		{Name: "Transfer In", Type: "transfer"},
		{Name: "Transfer Out", Type: "transfer"},
	}
}

func (s *categoryService) validateCategory(category *entity.Category) error {
	if category == nil {
		return errors.ErrInvalidCategoryData
	}

	var validationErrors []string

	if category.UserID == uuid.Nil {
		validationErrors = append(validationErrors, "user ID is required")
	}
	if category.Name == "" {
		validationErrors = append(validationErrors, "name is required")
	}
	if category.Type == "" {
		validationErrors = append(validationErrors, "type is required")
	}

	if len(validationErrors) > 0 {
		return fmt.Errorf("%w: %v", errors.ErrValidation, validationErrors)
	}

	return nil
}

func (s *categoryService) buildCategoryTree(categories []entity.Category) []entity.CategoryTree {
	// Create a map for quick lookup
	categoryMap := make(map[uuid.UUID]entity.Category)
	for _, category := range categories {
		categoryMap[category.ID] = category
	}

	// Build tree
	var rootCategories []entity.CategoryTree
	for _, category := range categories {
		if category.ParentID == nil {
			tree := s.buildSubtree(category, categoryMap)
			rootCategories = append(rootCategories, tree)
		}
	}

	return rootCategories
}

func (s *categoryService) buildSubtree(category entity.Category, categoryMap map[uuid.UUID]entity.Category) entity.CategoryTree {
	tree := entity.CategoryTree{
		Category: category,
	}

	// Find children
	for _, potentialChild := range categoryMap {
		if potentialChild.ParentID != nil && *potentialChild.ParentID == category.ID {
			childTree := s.buildSubtree(potentialChild, categoryMap)
			tree.Children = append(tree.Children, childTree)
		}
	}

	return tree
}

func (s *categoryService) wouldCreateCircularReference(ctx context.Context, categoryID uuid.UUID, newParentID uuid.UUID) bool {
	// If the new parent is the same as the category, it's circular
	if categoryID == newParentID {
		return true
	}

	// Get the potential parent
	parent, err := s.categoryRepo.GetByID(ctx, newParentID)
	if err != nil || parent == nil {
		return false
	}

	// Check if any of the parent's ancestors is the category we're trying to move
	current := parent
	for current.ParentID != nil {
		if *current.ParentID == categoryID {
			return true
		}
		current, err = s.categoryRepo.GetByID(ctx, *current.ParentID)
		if err != nil || current == nil {
			break
		}
	}

	return false
}
