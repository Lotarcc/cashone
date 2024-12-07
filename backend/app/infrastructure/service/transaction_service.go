package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"cashone/domain/entity"
	"cashone/domain/repository"
	"cashone/domain/service"
)

var (
	ErrTransactionNotFound = errors.New("transaction not found")
	ErrInvalidTransaction  = errors.New("invalid transaction data")
	ErrInsufficientFunds   = errors.New("insufficient funds")
)

type transactionService struct {
	transactionRepo repository.TransactionRepository
	cardRepo        repository.CardRepository
	categoryRepo    repository.CategoryRepository
	log             *zap.SugaredLogger
}

func NewTransactionService(
	transactionRepo repository.TransactionRepository,
	cardRepo repository.CardRepository,
	categoryRepo repository.CategoryRepository,
	log *zap.SugaredLogger,
) service.TransactionService {
	return &transactionService{
		transactionRepo: transactionRepo,
		cardRepo:        cardRepo,
		categoryRepo:    categoryRepo,
		log:             log,
	}
}

func (s *transactionService) CreateTransaction(ctx context.Context, userID uuid.UUID, transaction *entity.Transaction) error {
	// Verify card exists and belongs to user
	card, err := s.cardRepo.GetByID(ctx, transaction.CardID)
	if err != nil {
		s.log.Errorw("Failed to get card for transaction", "error", err, "card_id", transaction.CardID)
		return fmt.Errorf("failed to verify card: %w", err)
	}
	if card == nil {
		return ErrCardNotFound
	}
	if card.UserID != userID {
		return ErrUnauthorized
	}

	// Verify category if provided
	if transaction.CategoryID != nil {
		category, err := s.categoryRepo.GetByID(ctx, *transaction.CategoryID)
		if err != nil {
			s.log.Errorw("Failed to get category", "error", err, "category_id", transaction.CategoryID)
			return fmt.Errorf("failed to verify category: %w", err)
		}
		if category == nil {
			return fmt.Errorf("category not found")
		}
		if category.UserID != userID {
			return ErrUnauthorized
		}
		if category.Type != transaction.Type {
			return fmt.Errorf("category type does not match transaction type")
		}
	}

	// Validate transaction data
	if err := s.validateTransaction(transaction, card); err != nil {
		return err
	}

	// Set transaction properties
	transaction.UserID = userID
	if transaction.TransactionDate.IsZero() {
		transaction.TransactionDate = time.Now()
	}

	// Calculate balance after transaction
	newBalance := card.Balance
	switch transaction.Type {
	case "expense":
		newBalance -= transaction.Amount
	case "income":
		newBalance += transaction.Amount
	case "transfer":
		newBalance -= transaction.Amount
	}

	// Check if new balance would exceed credit limit
	if newBalance < -card.CreditLimit {
		return ErrInsufficientFunds
	}

	transaction.BalanceAfter = newBalance

	// Create the transaction
	if err := s.transactionRepo.Create(ctx, transaction); err != nil {
		s.log.Errorw("Failed to create transaction",
			"error", err,
			"user_id", userID,
			"card_id", transaction.CardID,
			"amount", transaction.Amount,
		)
		return fmt.Errorf("failed to create transaction: %w", err)
	}

	return nil
}

func (s *transactionService) GetCardTransactions(ctx context.Context, userID, cardID uuid.UUID, limit, offset int) ([]entity.Transaction, error) {
	// Verify card ownership
	card, err := s.cardRepo.GetByID(ctx, cardID)
	if err != nil {
		s.log.Errorw("Failed to get card", "error", err, "card_id", cardID)
		return nil, fmt.Errorf("failed to verify card: %w", err)
	}
	if card == nil {
		return nil, ErrCardNotFound
	}
	if card.UserID != userID {
		return nil, ErrUnauthorized
	}

	// Get transactions
	transactions, err := s.transactionRepo.GetByCardID(ctx, cardID, limit, offset)
	if err != nil {
		s.log.Errorw("Failed to get card transactions",
			"error", err,
			"card_id", cardID,
			"limit", limit,
			"offset", offset,
		)
		return nil, fmt.Errorf("failed to get transactions: %w", err)
	}

	return transactions, nil
}

func (s *transactionService) GetUserTransactions(ctx context.Context, userID uuid.UUID, limit, offset int) ([]entity.Transaction, error) {
	transactions, err := s.transactionRepo.GetByUserID(ctx, userID, limit, offset)
	if err != nil {
		s.log.Errorw("Failed to get user transactions",
			"error", err,
			"user_id", userID,
			"limit", limit,
			"offset", offset,
		)
		return nil, fmt.Errorf("failed to get transactions: %w", err)
	}

	return transactions, nil
}

func (s *transactionService) UpdateTransaction(ctx context.Context, userID uuid.UUID, transaction *entity.Transaction) error {
	// Get existing transaction
	existingTx, err := s.transactionRepo.GetByID(ctx, transaction.ID)
	if err != nil {
		s.log.Errorw("Failed to get transaction for update", "error", err, "transaction_id", transaction.ID)
		return fmt.Errorf("failed to get transaction: %w", err)
	}
	if existingTx == nil {
		return ErrTransactionNotFound
	}

	// Verify ownership
	if existingTx.UserID != userID {
		return ErrUnauthorized
	}

	// Prevent modification of Monobank transactions
	if existingTx.MonobankID != "" {
		return fmt.Errorf("cannot modify monobank transaction")
	}

	// Get card
	card, err := s.cardRepo.GetByID(ctx, existingTx.CardID)
	if err != nil {
		s.log.Errorw("Failed to get card", "error", err, "card_id", existingTx.CardID)
		return fmt.Errorf("failed to get card: %w", err)
	}
	if card == nil {
		return ErrCardNotFound
	}

	// Verify category if provided
	if transaction.CategoryID != nil {
		category, err := s.categoryRepo.GetByID(ctx, *transaction.CategoryID)
		if err != nil {
			s.log.Errorw("Failed to get category", "error", err, "category_id", transaction.CategoryID)
			return fmt.Errorf("failed to verify category: %w", err)
		}
		if category == nil {
			return fmt.Errorf("category not found")
		}
		if category.UserID != userID {
			return ErrUnauthorized
		}
		if category.Type != transaction.Type {
			return fmt.Errorf("category type does not match transaction type")
		}
	}

	// Validate updated transaction data
	if err := s.validateTransaction(transaction, card); err != nil {
		return err
	}

	// Preserve unchangeable fields
	transaction.UserID = existingTx.UserID
	transaction.CardID = existingTx.CardID
	transaction.MonobankID = ""

	// Update the transaction
	if err := s.transactionRepo.Update(ctx, transaction); err != nil {
		s.log.Errorw("Failed to update transaction",
			"error", err,
			"transaction_id", transaction.ID,
			"user_id", userID,
		)
		return fmt.Errorf("failed to update transaction: %w", err)
	}

	return nil
}

func (s *transactionService) DeleteTransaction(ctx context.Context, userID, transactionID uuid.UUID) error {
	// Get existing transaction
	transaction, err := s.transactionRepo.GetByID(ctx, transactionID)
	if err != nil {
		s.log.Errorw("Failed to get transaction for deletion", "error", err, "transaction_id", transactionID)
		return fmt.Errorf("failed to get transaction: %w", err)
	}
	if transaction == nil {
		return ErrTransactionNotFound
	}

	// Verify ownership
	if transaction.UserID != userID {
		return ErrUnauthorized
	}

	// Prevent deletion of Monobank transactions
	if transaction.MonobankID != "" {
		return fmt.Errorf("cannot delete monobank transaction")
	}

	// Delete the transaction
	if err := s.transactionRepo.Delete(ctx, transactionID); err != nil {
		s.log.Errorw("Failed to delete transaction",
			"error", err,
			"transaction_id", transactionID,
			"user_id", userID,
		)
		return fmt.Errorf("failed to delete transaction: %w", err)
	}

	return nil
}

func (s *transactionService) validateTransaction(transaction *entity.Transaction, card *entity.Card) error {
	if transaction.Amount <= 0 {
		return fmt.Errorf("%w: amount must be positive", ErrInvalidTransaction)
	}

	if transaction.Type != "expense" && transaction.Type != "income" && transaction.Type != "transfer" {
		return fmt.Errorf("%w: invalid transaction type", ErrInvalidTransaction)
	}

	if transaction.CurrencyCode != card.CurrencyCode {
		return fmt.Errorf("%w: transaction currency must match card currency", ErrInvalidTransaction)
	}

	if transaction.Description == "" {
		return fmt.Errorf("%w: description is required", ErrInvalidTransaction)
	}

	return nil
}
