package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"cashone/domain/entity"
	"cashone/domain/errors"
	"cashone/domain/repository"
)

// TransactionService handles transaction-related business logic
type TransactionService struct {
	transactionRepo repository.TransactionRepository
	log             *zap.SugaredLogger
}

// NewTransactionService creates a new transaction service instance
func NewTransactionService(
	transactionRepo repository.TransactionRepository,
	log *zap.SugaredLogger,
) *TransactionService {
	return &TransactionService{
		transactionRepo: transactionRepo,
		log:             log,
	}
}

// Create creates a new transaction
func (s *TransactionService) Create(ctx context.Context, transaction *entity.Transaction) error {
	return s.transactionRepo.Create(ctx, transaction)
}

// GetByID retrieves a transaction by its ID
func (s *TransactionService) GetByID(ctx context.Context, id uuid.UUID) (*entity.Transaction, error) {
	transaction, err := s.transactionRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errors.ErrDatabaseOperation, err)
	}
	if transaction == nil {
		return nil, errors.ErrTransactionNotFound
	}
	return transaction, nil
}

// GetByCardID retrieves transactions by card ID with pagination
func (s *TransactionService) GetByCardID(ctx context.Context, cardID uuid.UUID, limit, offset int) ([]entity.Transaction, error) {
	return s.transactionRepo.GetByCardID(ctx, cardID, limit, offset)
}

// GetByUserID retrieves transactions by user ID with pagination
func (s *TransactionService) GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]entity.Transaction, error) {
	return s.transactionRepo.GetByUserID(ctx, userID, limit, offset)
}

// Update updates an existing transaction
func (s *TransactionService) Update(ctx context.Context, transaction *entity.Transaction) error {
	return s.transactionRepo.Update(ctx, transaction)
}

// Delete deletes a transaction by its ID
func (s *TransactionService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.transactionRepo.Delete(ctx, id)
}

// Search searches for transactions with filters and pagination
func (s *TransactionService) Search(ctx context.Context, userID uuid.UUID, params entity.TransactionSearchParams, limit, offset int) ([]entity.Transaction, error) {
	return s.transactionRepo.Search(ctx, userID, params, limit, offset)
}
