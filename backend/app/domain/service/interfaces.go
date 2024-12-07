package service

import (
	"context"

	"github.com/google/uuid"

	"cashone/domain/entity"
)

type UserService interface {
	Register(ctx context.Context, email, password, name string) (*entity.User, error)
	Login(ctx context.Context, email, password string) (*entity.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error)
	UpdateProfile(ctx context.Context, id uuid.UUID, name string) error
	ChangePassword(ctx context.Context, id uuid.UUID, oldPassword, newPassword string) error
}

type CardService interface {
	CreateManualCard(ctx context.Context, userID uuid.UUID, card *entity.Card) error
	GetUserCards(ctx context.Context, userID uuid.UUID) ([]entity.Card, error)
	UpdateCard(ctx context.Context, userID uuid.UUID, card *entity.Card) error
	DeleteCard(ctx context.Context, userID, cardID uuid.UUID) error
}

type TransactionService interface {
	CreateTransaction(ctx context.Context, userID uuid.UUID, transaction *entity.Transaction) error
	GetCardTransactions(ctx context.Context, userID, cardID uuid.UUID, limit, offset int) ([]entity.Transaction, error)
	GetUserTransactions(ctx context.Context, userID uuid.UUID, limit, offset int) ([]entity.Transaction, error)
	UpdateTransaction(ctx context.Context, userID uuid.UUID, transaction *entity.Transaction) error
	DeleteTransaction(ctx context.Context, userID, transactionID uuid.UUID) error
}

type CategoryService interface {
	CreateCategory(ctx context.Context, userID uuid.UUID, category *entity.Category) error
	GetUserCategories(ctx context.Context, userID uuid.UUID) ([]entity.Category, error)
	UpdateCategory(ctx context.Context, userID uuid.UUID, category *entity.Category) error
	DeleteCategory(ctx context.Context, userID, categoryID uuid.UUID) error
}

type MonobankService interface {
	IntegrateMonobank(ctx context.Context, userID uuid.UUID, token string) error
	SyncMonobankData(ctx context.Context, userID uuid.UUID) error
	GetMonobankIntegration(ctx context.Context, userID uuid.UUID) (*entity.MonobankIntegration, error)
	DisconnectMonobank(ctx context.Context, userID uuid.UUID) error
	HandleWebhook(ctx context.Context, data []byte) error
}

// Request/Response types for specific operations

type TransactionFilter struct {
	StartDate   *string    `json:"start_date,omitempty"`
	EndDate     *string    `json:"end_date,omitempty"`
	CategoryID  *uuid.UUID `json:"category_id,omitempty"`
	Type        *string    `json:"type,omitempty"`
	MinAmount   *int64     `json:"min_amount,omitempty"`
	MaxAmount   *int64     `json:"max_amount,omitempty"`
	Description *string    `json:"description,omitempty"`
	SortBy      *string    `json:"sort_by,omitempty"`
	SortOrder   *string    `json:"sort_order,omitempty"`
}

type TransactionStats struct {
	TotalIncome     int64                  `json:"total_income"`
	TotalExpense    int64                  `json:"total_expense"`
	NetAmount       int64                  `json:"net_amount"`
	CategorySummary []CategoryTransactions `json:"category_summary"`
}

type CategoryTransactions struct {
	CategoryID   uuid.UUID `json:"category_id"`
	CategoryName string    `json:"category_name"`
	Amount       int64     `json:"amount"`
	Count        int       `json:"count"`
}
