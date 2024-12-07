package service

import (
	"go.uber.org/zap"

	"cashone/domain/repository"
	"cashone/domain/service"
)

// Factory provides an interface to create all services
type Factory interface {
	NewUserService() service.UserService
	NewCardService() service.CardService
	NewTransactionService() service.TransactionService
	NewCategoryService() service.CategoryService
	NewMonobankService() service.MonobankService
}

type factory struct {
	repoFactory repository.Factory
	log         *zap.SugaredLogger
}

// NewFactory creates a new service factory
func NewFactory(repoFactory repository.Factory, log *zap.SugaredLogger) Factory {
	return &factory{
		repoFactory: repoFactory,
		log:         log,
	}
}

func (f *factory) NewUserService() service.UserService {
	return NewUserService(
		f.repoFactory.NewUserRepository(),
		f.log,
	)
}

func (f *factory) NewCardService() service.CardService {
	return NewCardService(
		f.repoFactory.NewCardRepository(),
		f.repoFactory.NewUserRepository(),
		f.log,
	)
}

func (f *factory) NewTransactionService() service.TransactionService {
	return NewTransactionService(
		f.repoFactory.NewTransactionRepository(),
		f.repoFactory.NewCardRepository(),
		f.repoFactory.NewCategoryRepository(),
		f.log,
	)
}

func (f *factory) NewCategoryService() service.CategoryService {
	return NewCategoryService(
		f.repoFactory.NewCategoryRepository(),
		f.repoFactory.NewUserRepository(),
		f.log,
	)
}

func (f *factory) NewMonobankService() service.MonobankService {
	return NewMonobankService(
		f.repoFactory.NewMonobankIntegrationRepository(),
		f.repoFactory.NewCardRepository(),
		f.repoFactory.NewTransactionRepository(),
		f.repoFactory.NewUserRepository(),
		f.log,
	)
}
