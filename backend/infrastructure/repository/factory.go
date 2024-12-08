package repository

import (
	"go.uber.org/zap"
	"gorm.io/gorm"

	"cashone/domain/repository"
)

// Factory provides an interface to create all repositories
type Factory interface {
	NewUserRepository() repository.UserRepository
	NewCardRepository() repository.CardRepository
	NewTransactionRepository() repository.TransactionRepository
	NewCategoryRepository() repository.CategoryRepository
	NewMonobankIntegrationRepository() repository.MonobankIntegrationRepository
	NewRefreshTokenRepository() repository.RefreshTokenRepository
}

type factory struct {
	db  *gorm.DB
	log *zap.SugaredLogger
}

// NewFactory creates a new repository factory instance
func NewFactory(db *gorm.DB, log *zap.SugaredLogger) Factory {
	return &factory{
		db:  db,
		log: log,
	}
}

// NewUserRepository creates a new user repository instance
func (f *factory) NewUserRepository() repository.UserRepository {
	return NewUserRepository(f.db, f.log)
}

// NewCardRepository creates a new card repository instance
func (f *factory) NewCardRepository() repository.CardRepository {
	return NewCardRepository(f.db, f.log)
}

// NewTransactionRepository creates a new transaction repository instance
func (f *factory) NewTransactionRepository() repository.TransactionRepository {
	return NewTransactionRepository(f.db, f.log)
}

// NewCategoryRepository creates a new category repository instance
func (f *factory) NewCategoryRepository() repository.CategoryRepository {
	return NewCategoryRepository(f.db, f.log)
}

// NewMonobankIntegrationRepository creates a new Monobank integration repository instance
func (f *factory) NewMonobankIntegrationRepository() repository.MonobankIntegrationRepository {
	return NewMonobankIntegrationRepository(f.db, f.log)
}

// NewRefreshTokenRepository creates a new refresh token repository instance
func (f *factory) NewRefreshTokenRepository() repository.RefreshTokenRepository {
	return NewRefreshTokenRepository(f.db, f.log)
}
