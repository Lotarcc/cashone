package service

import (
	"context"
	"net/http"

	"github.com/google/uuid"

	"cashone/domain/entity"
)

// Factory provides an interface to create all services
type Factory interface {
	NewUserService() UserService
	NewCardService() CardService
	NewTransactionService() TransactionService
	NewCategoryService() CategoryService
	NewMonobankService() MonobankService
	NewAuthService() AuthService
}

// UserService handles user-related business logic
type UserService interface {
	Create(ctx context.Context, user *entity.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error)
	GetByEmail(ctx context.Context, email string) (*entity.User, error)
	Update(ctx context.Context, user *entity.User) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// CardService handles card-related business logic
type CardService interface {
	Create(ctx context.Context, card *entity.Card) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Card, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]entity.Card, error)
	Update(ctx context.Context, card *entity.Card) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// TransactionService handles transaction-related business logic
type TransactionService interface {
	Create(ctx context.Context, transaction *entity.Transaction) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Transaction, error)
	GetByCardID(ctx context.Context, cardID uuid.UUID, limit, offset int) ([]entity.Transaction, error)
	GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]entity.Transaction, error)
	Update(ctx context.Context, transaction *entity.Transaction) error
	Delete(ctx context.Context, id uuid.UUID) error
	Search(ctx context.Context, userID uuid.UUID, params entity.TransactionSearchParams, limit, offset int) ([]entity.Transaction, error)
}

// CategoryService handles category-related business logic
type CategoryService interface {
	Create(ctx context.Context, category *entity.Category) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Category, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]entity.Category, error)
	Update(ctx context.Context, category *entity.Category) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetTree(ctx context.Context, userID uuid.UUID) ([]entity.CategoryTree, error)
	GetChildren(ctx context.Context, categoryID uuid.UUID) ([]entity.Category, error)
	MoveCategory(ctx context.Context, categoryID uuid.UUID, newParentID *uuid.UUID) error
	CreateDefaultCategories(ctx context.Context, userID uuid.UUID) error
	GetDefaultCategories() []entity.Category
}

// MonobankService defines the interface for Monobank integration operations
type MonobankService interface {
	Connect(ctx context.Context, userID uuid.UUID, token string) error
	Disconnect(ctx context.Context, userID uuid.UUID) error
	SyncUserData(ctx context.Context, userID uuid.UUID) error
	HandleWebhook(ctx context.Context, data []byte) error
	GetStatus(ctx context.Context, userID uuid.UUID) (*entity.MonobankIntegration, error)
	SetHTTPClient(client interface {
		Do(*http.Request) (*http.Response, error)
	})
}

// AuthService handles authentication-related business logic
type AuthService interface {
	Register(ctx context.Context, req *entity.RegisterRequest) (*entity.RegisterResponse, error)
	Login(ctx context.Context, req *entity.LoginRequest) (*entity.LoginResponse, error)
	RefreshToken(ctx context.Context, token string) (*entity.AuthToken, error)
	Logout(ctx context.Context, userID uuid.UUID, token string) error
	ValidateToken(ctx context.Context, token string) (*entity.Claims, error)
	HashPassword(password string) (string, error)
	VerifyPassword(password, hash string) error
	GenerateTokens(ctx context.Context, user *entity.User, userAgent, ip string) (*entity.AuthToken, error)
	RevokeAllUserTokens(ctx context.Context, userID uuid.UUID) error
	GetActiveTokens(ctx context.Context, userID uuid.UUID) ([]entity.RefreshToken, error)
}
