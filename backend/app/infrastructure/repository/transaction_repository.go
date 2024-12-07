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

type transactionRepository struct {
	db  *gorm.DB
	log *zap.SugaredLogger
}

func NewTransactionRepository(db *gorm.DB, log *zap.SugaredLogger) repository.TransactionRepository {
	return &transactionRepository{
		db:  db,
		log: log,
	}
}

func (r *transactionRepository) Create(ctx context.Context, transaction *entity.Transaction) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Create the transaction
		if err := tx.Create(transaction).Error; err != nil {
			r.log.Errorw("Failed to create transaction",
				"error", err,
				"user_id", transaction.UserID,
				"card_id", transaction.CardID,
				"amount", transaction.Amount,
			)
			return err
		}

		// Update card balance
		var card entity.Card
		if err := tx.First(&card, transaction.CardID).Error; err != nil {
			r.log.Errorw("Failed to get card for balance update",
				"error", err,
				"card_id", transaction.CardID,
			)
			return err
		}

		// Update card balance based on transaction type
		switch transaction.Type {
		case "income":
			card.Balance += transaction.Amount
		case "expense":
			card.Balance -= transaction.Amount
		case "transfer":
			// For transfers, the balance update depends on whether this is the source or destination card
			// This might need additional logic depending on your transfer implementation
			card.Balance -= transaction.Amount
		}

		if err := tx.Save(&card).Error; err != nil {
			r.log.Errorw("Failed to update card balance",
				"error", err,
				"card_id", card.ID,
				"new_balance", card.Balance,
			)
			return err
		}

		return nil
	})
}

func (r *transactionRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Transaction, error) {
	var transaction entity.Transaction
	if err := r.db.WithContext(ctx).First(&transaction, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		r.log.Errorw("Failed to get transaction by ID", "error", err, "id", id)
		return nil, err
	}
	return &transaction, nil
}

func (r *transactionRepository) GetByCardID(ctx context.Context, cardID uuid.UUID, limit, offset int) ([]entity.Transaction, error) {
	var transactions []entity.Transaction
	query := r.db.WithContext(ctx).
		Where("card_id = ?", cardID).
		Order("transaction_date DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&transactions).Error; err != nil {
		r.log.Errorw("Failed to get transactions by card ID",
			"error", err,
			"card_id", cardID,
			"limit", limit,
			"offset", offset,
		)
		return nil, err
	}
	return transactions, nil
}

func (r *transactionRepository) GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]entity.Transaction, error) {
	var transactions []entity.Transaction
	query := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("transaction_date DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&transactions).Error; err != nil {
		r.log.Errorw("Failed to get transactions by user ID",
			"error", err,
			"user_id", userID,
			"limit", limit,
			"offset", offset,
		)
		return nil, err
	}
	return transactions, nil
}

func (r *transactionRepository) GetByMonobankID(ctx context.Context, monobankID string) (*entity.Transaction, error) {
	var transaction entity.Transaction
	if err := r.db.WithContext(ctx).
		Where("monobank_id = ?", monobankID).
		First(&transaction).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		r.log.Errorw("Failed to get transaction by Monobank ID",
			"error", err,
			"monobank_id", monobankID,
		)
		return nil, err
	}
	return &transaction, nil
}

func (r *transactionRepository) Update(ctx context.Context, transaction *entity.Transaction) error {
	result := r.db.WithContext(ctx).Model(transaction).Updates(map[string]interface{}{
		"category_id":             transaction.CategoryID,
		"amount":                  transaction.Amount,
		"operation_amount":        transaction.OperationAmount,
		"currency_code":           transaction.CurrencyCode,
		"operation_currency_code": transaction.OperationCurrencyCode,
		"type":                    transaction.Type,
		"description":             transaction.Description,
		"mcc":                     transaction.MCC,
		"commission_rate":         transaction.CommissionRate,
		"cashback_amount":         transaction.CashbackAmount,
		"balance_after":           transaction.BalanceAfter,
		"hold":                    transaction.Hold,
		"transaction_date":        transaction.TransactionDate,
		"comment":                 transaction.Comment,
	})

	if result.Error != nil {
		r.log.Errorw("Failed to update transaction",
			"error", result.Error,
			"id", transaction.ID,
		)
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *transactionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&entity.Transaction{}, "id = ?", id)
	if result.Error != nil {
		r.log.Errorw("Failed to delete transaction", "error", result.Error, "id", id)
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}
