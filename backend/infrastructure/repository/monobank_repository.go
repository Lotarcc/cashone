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

type monobankIntegrationRepository struct {
	db  *gorm.DB
	log *zap.SugaredLogger
}

// NewMonobankIntegrationRepository creates a new Monobank integration repository instance
func NewMonobankIntegrationRepository(db *gorm.DB, log *zap.SugaredLogger) repository.MonobankIntegrationRepository {
	return &monobankIntegrationRepository{
		db:  db,
		log: log,
	}
}

func (r *monobankIntegrationRepository) Create(ctx context.Context, integration *entity.MonobankIntegration) error {
	// Check if integration already exists for this user
	var existing entity.MonobankIntegration
	err := r.db.WithContext(ctx).
		Where("user_id = ?", integration.UserID).
		First(&existing).Error

	if err == nil {
		r.log.Warnw("Monobank integration already exists for user",
			"user_id", integration.UserID,
		)
		return errors.New("monobank integration already exists for this user")
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		r.log.Errorw("Error checking existing monobank integration",
			"error", err,
			"user_id", integration.UserID,
		)
		return err
	}

	// Create new integration
	if err := r.db.WithContext(ctx).Create(integration).Error; err != nil {
		r.log.Errorw("Failed to create monobank integration",
			"error", err,
			"user_id", integration.UserID,
		)
		return err
	}

	return nil
}

func (r *monobankIntegrationRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*entity.MonobankIntegration, error) {
	var integration entity.MonobankIntegration
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		First(&integration).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		r.log.Errorw("Failed to get monobank integration",
			"error", err,
			"user_id", userID,
		)
		return nil, err
	}
	return &integration, nil
}

func (r *monobankIntegrationRepository) Update(ctx context.Context, integration *entity.MonobankIntegration) error {
	result := r.db.WithContext(ctx).Model(integration).Updates(map[string]interface{}{
		"token":       integration.Token,
		"webhook_url": integration.WebhookURL,
		"permissions": integration.Permissions,
	})

	if result.Error != nil {
		r.log.Errorw("Failed to update monobank integration",
			"error", result.Error,
			"user_id", integration.UserID,
		)
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *monobankIntegrationRepository) Delete(ctx context.Context, userID uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// First, get all cards associated with this integration
		var cards []entity.Card
		if err := tx.Where("user_id = ? AND is_manual = false", userID).Find(&cards).Error; err != nil {
			r.log.Errorw("Failed to get monobank cards",
				"error", err,
				"user_id", userID,
			)
			return err
		}

		// Delete all transactions for these cards
		for _, card := range cards {
			if err := tx.Where("card_id = ?", card.ID).Delete(&entity.Transaction{}).Error; err != nil {
				r.log.Errorw("Failed to delete card transactions",
					"error", err,
					"card_id", card.ID,
				)
				return err
			}
		}

		// Delete the cards
		if err := tx.Where("user_id = ? AND is_manual = false", userID).Delete(&entity.Card{}).Error; err != nil {
			r.log.Errorw("Failed to delete monobank cards",
				"error", err,
				"user_id", userID,
			)
			return err
		}

		// Finally, delete the integration
		result := tx.Delete(&entity.MonobankIntegration{}, "user_id = ?", userID)
		if result.Error != nil {
			r.log.Errorw("Failed to delete monobank integration",
				"error", result.Error,
				"user_id", userID,
			)
			return result.Error
		}

		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}

		return nil
	})
}
