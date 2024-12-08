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

type cardRepository struct {
	db  *gorm.DB
	log *zap.SugaredLogger
}

// NewCardRepository creates a new card repository instance
func NewCardRepository(db *gorm.DB, log *zap.SugaredLogger) repository.CardRepository {
	return &cardRepository{
		db:  db,
		log: log,
	}
}

func (r *cardRepository) Create(ctx context.Context, card *entity.Card) error {
	if err := r.db.WithContext(ctx).Create(card).Error; err != nil {
		r.log.Errorw("Failed to create card",
			"error", err,
			"user_id", card.UserID,
			"card_name", card.CardName,
			"is_manual", card.IsManual,
		)
		return err
	}
	return nil
}

func (r *cardRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Card, error) {
	var card entity.Card
	if err := r.db.WithContext(ctx).First(&card, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		r.log.Errorw("Failed to get card by ID", "error", err, "id", id)
		return nil, err
	}
	return &card, nil
}

func (r *cardRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]entity.Card, error) {
	var cards []entity.Card
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&cards).Error; err != nil {
		r.log.Errorw("Failed to get cards by user ID", "error", err, "user_id", userID)
		return nil, err
	}
	return cards, nil
}

func (r *cardRepository) GetByMonobankAccountID(ctx context.Context, accountID string) (*entity.Card, error) {
	var card entity.Card
	if err := r.db.WithContext(ctx).
		Where("monobank_account_id = ? AND is_manual = false", accountID).
		First(&card).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		r.log.Errorw("Failed to get card by Monobank account ID",
			"error", err,
			"monobank_account_id", accountID,
		)
		return nil, err
	}
	return &card, nil
}

func (r *cardRepository) Update(ctx context.Context, card *entity.Card) error {
	result := r.db.WithContext(ctx).Model(card).Updates(map[string]interface{}{
		"card_name":           card.CardName,
		"masked_pan":          card.MaskedPan,
		"balance":             card.Balance,
		"credit_limit":        card.CreditLimit,
		"currency_code":       card.CurrencyCode,
		"type":                card.Type,
		"monobank_account_id": card.MonobankAccountID,
	})

	if result.Error != nil {
		r.log.Errorw("Failed to update card",
			"error", result.Error,
			"id", card.ID,
			"user_id", card.UserID,
		)
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *cardRepository) Delete(ctx context.Context, id uuid.UUID) error {
	// Start a transaction to handle cascading deletes
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// First delete associated transactions
		if err := tx.Where("card_id = ?", id).Delete(&entity.Transaction{}).Error; err != nil {
			r.log.Errorw("Failed to delete card's transactions", "error", err, "card_id", id)
			return err
		}

		// Then delete the card
		result := tx.Delete(&entity.Card{}, "id = ?", id)
		if result.Error != nil {
			r.log.Errorw("Failed to delete card", "error", result.Error, "id", id)
			return result.Error
		}

		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}

		return nil
	})
}
