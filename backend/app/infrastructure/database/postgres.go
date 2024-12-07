package database

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"cashone/domain/entity"
)

func NewPostgresDB(log *zap.SugaredLogger) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=UTC",
		viper.GetString("database.host"),
		viper.GetString("database.user"),
		viper.GetString("database.password"),
		viper.GetString("database.name"),
		viper.GetInt("database.port"),
	)

	config := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}

	db, err := gorm.Open(postgres.Open(dsn), config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	// Set connection pool settings
	sqlDB.SetMaxOpenConns(viper.GetInt("database.max_open_conns"))
	sqlDB.SetMaxIdleConns(viper.GetInt("database.max_idle_conns"))
	sqlDB.SetConnMaxLifetime(time.Duration(viper.GetInt("database.conn_max_lifetime")) * time.Second)

	// Auto-migrate the schema
	if err := db.AutoMigrate(
		&entity.User{},
		&entity.Card{},
		&entity.Transaction{},
		&entity.Category{},
		&entity.MonobankIntegration{},
	); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	log.Info("Successfully connected to database")
	return db, nil
}
