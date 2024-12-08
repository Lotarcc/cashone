package repository

import (
	"context"

	"github.com/google/uuid"

	"cashone/domain/entity"
)

// Factory provides an interface to create all repositories
type Factory interface {
	NewUserRepository() UserRepository
	NewCardRepository() CardRepository
	NewTransactionRepository() TransactionRepository
	NewCategoryRepository() CategoryRepository
	NewMonobankIntegrationRepository() MonobankIntegrationRepository
	NewRefreshTokenRepository() RefreshTokenRepository
}

// UserRepository defines the interface for user-related database operations
type UserRepository interface {
	Create(ctx context.Context, user *entity.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error)
	GetByEmail(ctx context.Context, email string) (*entity.User, error)
	Update(ctx context.Context, user *entity.User) error
	Delete(ctx context.Context, id uuid.UUID) error
	Ping(ctx context.Context) error
}

// CardRepository defines the interface for card-related database operations
type CardRepository interface {
	Create(ctx context.Context, card *entity.Card) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Card, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]entity.Card, error)
	GetByMonobankAccountID(ctx context.Context, accountID string) (*entity.Card, error)
	Update(ctx context.Context, card *entity.Card) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// TransactionRepository defines the interface for transaction-related database operations
type TransactionRepository interface {
	Create(ctx context.Context, transaction *entity.Transaction) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Transaction, error)
	GetByCardID(ctx context.Context, cardID uuid.UUID, limit, offset int) ([]entity.Transaction, error)
	GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]entity.Transaction, error)
	GetByMonobankID(ctx context.Context, monobankID string) (*entity.Transaction, error)
	Update(ctx context.Context, transaction *entity.Transaction) error
	Delete(ctx context.Context, id uuid.UUID) error
	Search(ctx context.Context, userID uuid.UUID, params entity.TransactionSearchParams, limit, offset int) ([]entity.Transaction, error)
}

// CategoryRepository defines the interface for category-related database operations
type CategoryRepository interface {
	Create(ctx context.Context, category *entity.Category) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Category, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]entity.Category, error)
	Update(ctx context.Context, category *entity.Category) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// MonobankIntegrationRepository defines the interface for Monobank integration-related database operations
type MonobankIntegrationRepository interface {
	Create(ctx context.Context, integration *entity.MonobankIntegration) error
	GetByUserID(ctx context.Context, userID uuid.UUID) (*entity.MonobankIntegration, error)
	Update(ctx context.Context, integration *entity.MonobankIntegration) error
	Delete(ctx context.Context, userID uuid.UUID) error
}

// RefreshTokenRepository defines the interface for refresh token-related database operations
type RefreshTokenRepository interface {
	Create(ctx context.Context, token *entity.RefreshToken) error
	GetByToken(ctx context.Context, token string) (*entity.RefreshToken, error)
	GetActiveByUserID(ctx context.Context, userID uuid.UUID) ([]entity.RefreshToken, error)
	Revoke(ctx context.Context, token string) error
	RevokeAllUserTokens(ctx context.Context, userID uuid.UUID) error
	DeleteExpired(ctx context.Context) error
	Update(ctx context.Context, token *entity.RefreshToken) error
}
