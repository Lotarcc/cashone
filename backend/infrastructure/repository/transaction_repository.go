package repository

import (
	"context"
	"fmt"

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

// NewTransactionRepository creates a new transaction repository instance
func NewTransactionRepository(db *gorm.DB, log *zap.SugaredLogger) repository.TransactionRepository {
	return &transactionRepository{
		db:  db,
		log: log,
	}
}

func (r *transactionRepository) Create(ctx context.Context, transaction *entity.Transaction) error {
	return r.db.WithContext(ctx).Create(transaction).Error
}

func (r *transactionRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Transaction, error) {
	var transaction entity.Transaction
	err := r.db.WithContext(ctx).First(&transaction, "id = ?", id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &transaction, nil
}

func (r *transactionRepository) GetByCardID(ctx context.Context, cardID uuid.UUID, limit, offset int) ([]entity.Transaction, error) {
	var transactions []entity.Transaction
	err := r.db.WithContext(ctx).
		Where("card_id = ?", cardID).
		Order("transaction_date DESC").
		Limit(limit).
		Offset(offset).
		Find(&transactions).Error
	if err != nil {
		return nil, err
	}
	return transactions, nil
}

func (r *transactionRepository) GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]entity.Transaction, error) {
	var transactions []entity.Transaction
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("transaction_date DESC").
		Limit(limit).
		Offset(offset).
		Find(&transactions).Error
	if err != nil {
		return nil, err
	}
	return transactions, nil
}

func (r *transactionRepository) GetByMonobankID(ctx context.Context, monobankID string) (*entity.Transaction, error) {
	var transaction entity.Transaction
	err := r.db.WithContext(ctx).First(&transaction, "monobank_id = ?", monobankID).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &transaction, nil
}

func (r *transactionRepository) Update(ctx context.Context, transaction *entity.Transaction) error {
	return r.db.WithContext(ctx).Save(transaction).Error
}

func (r *transactionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entity.Transaction{}, "id = ?", id).Error
}

func (r *transactionRepository) Search(ctx context.Context, userID uuid.UUID, params entity.TransactionSearchParams, limit, offset int) ([]entity.Transaction, error) {
	query := r.db.WithContext(ctx).Model(&entity.Transaction{}).Where("user_id = ?", userID)

	// Apply filters
	if params.Query != "" {
		query = query.Where("description ILIKE ?", fmt.Sprintf("%%%s%%", params.Query))
	}

	if params.Type != "" {
		query = query.Where("type = ?", params.Type)
	}

	if params.CategoryID != nil {
		query = query.Where("category_id = ?", params.CategoryID)
	}

	if params.CardID != nil {
		query = query.Where("card_id = ?", params.CardID)
	}

	if params.FromDate != nil {
		query = query.Where("transaction_date >= ?", params.FromDate)
	}

	if params.ToDate != nil {
		query = query.Where("transaction_date <= ?", params.ToDate)
	}

	if params.MinAmount != nil {
		query = query.Where("amount >= ?", params.MinAmount)
	}

	if params.MaxAmount != nil {
		query = query.Where("amount <= ?", params.MaxAmount)
	}

	// Order by transaction date descending
	query = query.Order("transaction_date DESC")

	// Apply pagination
	query = query.Limit(limit).Offset(offset)

	var transactions []entity.Transaction
	if err := query.Find(&transactions).Error; err != nil {
		return nil, err
	}

	return transactions, nil
}
