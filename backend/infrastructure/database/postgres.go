package database

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"cashone/domain/entity"
	"cashone/pkg/config"
)

// DB represents a database connection
type DB struct {
	gorm   *gorm.DB
	logger *zap.SugaredLogger
}

// New creates a new database connection
func New(cfg *config.DatabaseConfig) (*DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host,
		cfg.Port,
		cfg.User,
		cfg.Password,
		cfg.Name,
		cfg.SSLMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB instance: %w", err)
	}

	// Set connection pool settings
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	return &DB{gorm: db}, nil
}

// NewPostgresDB creates a new database connection (for backward compatibility)
func NewPostgresDB(logger *zap.SugaredLogger, cfg *config.DatabaseConfig) (*DB, error) {
	db, err := New(cfg)
	if err != nil {
		return nil, err
	}
	db.logger = logger
	return db, nil
}

// GormDB returns the underlying gorm.DB instance
func (db *DB) GormDB() *gorm.DB {
	return db.gorm
}

// Close closes the database connection
func (db *DB) Close() error {
	sqlDB, err := db.gorm.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB instance: %w", err)
	}
	return sqlDB.Close()
}

// Ping checks the database connection
func (db *DB) Ping(ctx context.Context) error {
	sqlDB, err := db.gorm.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB instance: %w", err)
	}
	return sqlDB.PingContext(ctx)
}

// Truncate clears all tables in the database
func (db *DB) Truncate(ctx context.Context) error {
	tables := []string{
		"users",
		"categories",
		"cards",
		"transactions",
		"monobank_integrations",
		"refresh_tokens",
	}

	for _, table := range tables {
		if err := db.gorm.Exec(fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table)).Error; err != nil {
			return fmt.Errorf("failed to truncate table %s: %w", table, err)
		}
	}

	return nil
}

// Create creates a new record in the database
func (db *DB) Create(ctx context.Context, value interface{}) error {
	return db.gorm.WithContext(ctx).Create(value).Error
}

// GetByID retrieves a record by ID
func (db *DB) GetByID(ctx context.Context, id uuid.UUID) (interface{}, error) {
	// Get entity type from context
	entityType := ctx.Value("entity")
	if entityType == nil {
		return nil, fmt.Errorf("entity type not provided in context")
	}

	// Create a new instance of the same type
	var result interface{}
	switch v := entityType.(type) {
	case *entity.User:
		result = &entity.User{}
	case *entity.Category:
		result = &entity.Category{}
	case *entity.Card:
		result = &entity.Card{}
	case *entity.Transaction:
		result = &entity.Transaction{}
	case *entity.MonobankIntegration:
		result = &entity.MonobankIntegration{}
	case *entity.RefreshToken:
		result = &entity.RefreshToken{}
	default:
		return nil, fmt.Errorf("unsupported entity type: %T", v)
	}

	if err := db.gorm.WithContext(ctx).First(result, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	return result, nil
}

// Update updates a record in the database
func (db *DB) Update(ctx context.Context, value interface{}) error {
	return db.gorm.WithContext(ctx).Save(value).Error
}

// Delete deletes a record from the database
func (db *DB) Delete(ctx context.Context, value interface{}) error {
	return db.gorm.WithContext(ctx).Delete(value).Error
}
