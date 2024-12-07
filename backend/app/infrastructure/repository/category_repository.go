package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"cashone/domain/entity"
	"cashone/domain/repository"
)

type categoryRepository struct {
	db  *gorm.DB
	log *zap.SugaredLogger
}

func NewCategoryRepository(db *gorm.DB, log *zap.SugaredLogger) repository.CategoryRepository {
	return &categoryRepository{
		db:  db,
		log: log,
	}
}

func (r *categoryRepository) Create(ctx context.Context, category *entity.Category) error {
	if err := r.db.WithContext(ctx).Create(category).Error; err != nil {
		r.log.Errorw("Failed to create category",
			"error", err,
			"user_id", category.UserID,
			"name", category.Name,
			"type", category.Type,
		)
		return err
	}
	return nil
}

func (r *categoryRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Category, error) {
	var category entity.Category
	if err := r.db.WithContext(ctx).First(&category, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		r.log.Errorw("Failed to get category by ID", "error", err, "id", id)
		return nil, err
	}
	return &category, nil
}

func (r *categoryRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]entity.Category, error) {
	var categories []entity.Category
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("CASE WHEN parent_id IS NULL THEN 0 ELSE 1 END, name").
		Find(&categories).Error; err != nil {
		r.log.Errorw("Failed to get categories by user ID",
			"error", err,
			"user_id", userID,
		)
		return nil, err
	}
	return categories, nil
}

func (r *categoryRepository) Update(ctx context.Context, category *entity.Category) error {
	// Check for circular reference in parent_id if it exists
	if category.ParentID != nil {
		var parent entity.Category
		if err := r.db.First(&parent, "id = ?", category.ParentID).Error; err != nil {
			r.log.Errorw("Failed to get parent category",
				"error", err,
				"parent_id", category.ParentID,
			)
			return err
		}

		// Check if the parent category belongs to the same user
		if parent.UserID != category.UserID {
			r.log.Errorw("Attempted to set parent category from different user",
				"category_user_id", category.UserID,
				"parent_user_id", parent.UserID,
			)
			return errors.New("parent category must belong to the same user")
		}

		// Check if setting parent would create a cycle
		if err := r.checkCategoryCircularReference(ctx, category.ID, *category.ParentID); err != nil {
			return err
		}
	}

	result := r.db.WithContext(ctx).Model(category).Updates(map[string]interface{}{
		"name":      category.Name,
		"parent_id": category.ParentID,
		"type":      category.Type,
	})

	if result.Error != nil {
		r.log.Errorw("Failed to update category",
			"error", result.Error,
			"id", category.ID,
		)
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *categoryRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Update child categories to remove parent reference
		if err := tx.Model(&entity.Category{}).
			Where("parent_id = ?", id).
			Update("parent_id", nil).Error; err != nil {
			r.log.Errorw("Failed to update child categories",
				"error", err,
				"parent_id", id,
			)
			return err
		}

		// Delete the category
		result := tx.Delete(&entity.Category{}, "id = ?", id)
		if result.Error != nil {
			r.log.Errorw("Failed to delete category", "error", result.Error, "id", id)
			return result.Error
		}

		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}

		return nil
	})
}

// checkCategoryCircularReference checks if setting parentID as the parent of categoryID
// would create a circular reference in the category hierarchy
func (r *categoryRepository) checkCategoryCircularReference(ctx context.Context, categoryID, parentID uuid.UUID) error {
	visited := make(map[uuid.UUID]bool)
	current := parentID

	for {
		if current == categoryID {
			r.log.Errorw("Circular reference detected in category hierarchy",
				"category_id", categoryID,
				"parent_id", parentID,
			)
			return errors.New("circular reference detected in category hierarchy")
		}

		visited[current] = true

		var parent entity.Category
		if err := r.db.WithContext(ctx).
			Select("id, parent_id").
			First(&parent, "id = ?", current).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				break
			}
			return err
		}

		if parent.ParentID == nil {
			break
		}

		current = *parent.ParentID
		if visited[current] {
			r.log.Errorw("Circular reference detected in existing category hierarchy",
				"category_id", categoryID,
				"parent_id", parentID,
			)
			return errors.New("circular reference detected in existing category hierarchy")
		}
	}

	return nil
}
