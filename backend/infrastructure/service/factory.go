package service

import (
	"go.uber.org/zap"

	"cashone/domain/repository"
	"cashone/domain/service"
	"cashone/pkg/config"
)

// serviceFactory implements the service.Factory interface
type serviceFactory struct {
	repoFactory repository.Factory
	config      *config.Config
	log         *zap.SugaredLogger
}

// NewFactory creates a new service factory instance
func NewFactory(repoFactory repository.Factory, config *config.Config, log *zap.SugaredLogger) service.Factory {
	return &serviceFactory{
		repoFactory: repoFactory,
		config:      config,
		log:         log,
	}
}

// NewUserService creates a new user service instance
func (f *serviceFactory) NewUserService() service.UserService {
	return NewUserService(f.repoFactory.NewUserRepository(), f.log)
}

// NewCardService creates a new card service instance
func (f *serviceFactory) NewCardService() service.CardService {
	return NewCardService(f.repoFactory.NewCardRepository(), f.repoFactory.NewUserRepository(), f.log)
}

// NewTransactionService creates a new transaction service instance
func (f *serviceFactory) NewTransactionService() service.TransactionService {
	return NewTransactionService(f.repoFactory.NewTransactionRepository(), f.log)
}

// NewCategoryService creates a new category service instance
func (f *serviceFactory) NewCategoryService() service.CategoryService {
	return NewCategoryService(f.repoFactory.NewCategoryRepository(), f.repoFactory.NewUserRepository(), f.log)
}

// NewMonobankService creates a new Monobank service instance
func (f *serviceFactory) NewMonobankService() service.MonobankService {
	return NewMonobankService(
		f.repoFactory.NewMonobankIntegrationRepository(),
		f.repoFactory.NewCardRepository(),
		f.repoFactory.NewTransactionRepository(),
		f.repoFactory.NewUserRepository(),
		f.log,
	)
}

// NewAuthService creates a new authentication service instance
func (f *serviceFactory) NewAuthService() service.AuthService {
	return NewAuthService(
		f.repoFactory.NewUserRepository(),
		f.repoFactory.NewRefreshTokenRepository(),
		f.config,
		f.log,
	)
}
