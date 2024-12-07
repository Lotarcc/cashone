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
}

type factory struct {
	db  *gorm.DB
	log *zap.SugaredLogger
}

// NewFactory creates a new repository factory
func NewFactory(db *gorm.DB, log *zap.SugaredLogger) Factory {
	return &factory{
		db:  db,
		log: log,
	}
}

func (f *factory) NewUserRepository() repository.UserRepository {
	return NewUserRepository(f.db, f.log)
}

func (f *factory) NewCardRepository() repository.CardRepository {
	return NewCardRepository(f.db, f.log)
}

func (f *factory) NewTransactionRepository() repository.TransactionRepository {
	return NewTransactionRepository(f.db, f.log)
}

func (f *factory) NewCategoryRepository() repository.CategoryRepository {
	return NewCategoryRepository(f.db, f.log)
}

func (f *factory) NewMonobankIntegrationRepository() repository.MonobankIntegrationRepository {
	return NewMonobankIntegrationRepository(f.db, f.log)
}
