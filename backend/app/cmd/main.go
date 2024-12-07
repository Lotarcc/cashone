package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/spf13/viper"
	echoSwagger "github.com/swaggo/echo-swagger"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"cashone/domain/repository"
	"cashone/infrastructure/database"
	"cashone/infrastructure/handler"
	infrarepo "cashone/infrastructure/repository"
	"cashone/infrastructure/service"
)

func initConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")

	// Read from environment variables
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			panic(fmt.Errorf("fatal error reading config file: %w", err))
		}
	}
}

func initLogger() *zap.Logger {
	var logger *zap.Logger
	var err error

	if viper.GetString("env") == "production" {
		logger, err = zap.NewProduction()
	} else {
		logger, err = zap.NewDevelopment()
	}

	if err != nil {
		panic(fmt.Errorf("failed to initialize logger: %w", err))
	}

	return logger
}

func setupEcho(log *zap.SugaredLogger) *echo.Echo {
	e := echo.New()

	// Middleware
	e.Use(middleware.RequestID())
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())
	e.Use(middleware.SecureWithConfig(middleware.SecureConfig{
		XSSProtection:         "1; mode=block",
		ContentTypeNosniff:    "nosniff",
		XFrameOptions:         "SAMEORIGIN",
		HSTSMaxAge:            31536000,
		HSTSExcludeSubdomains: false,
	}))

	// Swagger documentation
	if viper.GetString("env") != "production" {
		e.GET("/swagger/*", echoSwagger.WrapHandler)
	}

	return e
}

func initDependencies(db *gorm.DB, log *zap.SugaredLogger) (repository.Factory, service.Factory) {
	repoFactory := infrarepo.NewFactory(db, log)
	serviceFactory := service.NewFactory(repoFactory, log)
	return repoFactory, serviceFactory
}

func main() {
	// Initialize configuration
	initConfig()

	// Initialize logger
	logger := initLogger()
	defer logger.Sync()
	sugar := logger.Sugar()

	// Connect to database
	db, err := database.NewPostgresDB(sugar)
	if err != nil {
		sugar.Fatalw("Failed to connect to database", "error", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		sugar.Fatalw("Failed to get database instance", "error", err)
	}
	defer sqlDB.Close()

	// Initialize dependencies
	_, _ = initDependencies(db, sugar)

	// Initialize Echo
	e := setupEcho(sugar)

	// Initialize handlers
	handler.NewHealthHandler(e, sugar)

	// Start server
	go func() {
		if err := e.Start(fmt.Sprintf(":%s", viper.GetString("server.port"))); err != nil && err != http.ErrServerClosed {
			sugar.Fatalw("Failed to start server", "error", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		sugar.Fatalw("Failed to shutdown server gracefully", "error", err)
	}

	sugar.Info("Server stopped gracefully")
}
