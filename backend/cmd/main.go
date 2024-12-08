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
	echoSwagger "github.com/swaggo/echo-swagger"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gorm.io/gorm"

	_ "cashone/docs"
	"cashone/domain/repository"
	"cashone/domain/service"
	"cashone/infrastructure/database"
	"cashone/infrastructure/handler"
	authMiddleware "cashone/infrastructure/middleware"
	infrarepo "cashone/infrastructure/repository"
	infraservice "cashone/infrastructure/service"
	"cashone/pkg/config"
)

func initLogger(cfg *config.LoggerConfig) (*zap.Logger, error) {
	level := zap.NewAtomicLevel()
	if err := level.UnmarshalText([]byte(cfg.Level)); err != nil {
		return nil, fmt.Errorf("failed to parse log level: %w", err)
	}

	zapConfig := zap.Config{
		Level:            level,
		Development:      cfg.Encoding == "console",
		Encoding:         cfg.Encoding,
		OutputPaths:      cfg.OutputPaths,
		ErrorOutputPaths: cfg.ErrorOutputPaths,
		EncoderConfig:    zap.NewProductionEncoderConfig(),
	}

	zapConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	zapConfig.EncoderConfig.TimeKey = "timestamp"

	return zapConfig.Build()
}

func setupEcho(cfg *config.Config, log *zap.SugaredLogger) *echo.Echo {
	e := echo.New()

	// Middleware
	e.Use(middleware.RequestID())
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: `{"time":"${time_rfc3339_nano}","id":"${id}","remote_ip":"${remote_ip}",` +
			`"host":"${host}","method":"${method}","uri":"${uri}","user_agent":"${user_agent}",` +
			`"status":${status},"error":"${error}","latency":${latency},"latency_human":"${latency_human}"` +
			`,"bytes_in":${bytes_in},"bytes_out":${bytes_out}}` + "\n",
	}))
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     cfg.Server.CORS.AllowedOrigins,
		AllowMethods:     cfg.Server.CORS.AllowedMethods,
		AllowHeaders:     cfg.Server.CORS.AllowedHeaders,
		AllowCredentials: cfg.Server.CORS.AllowCredentials,
		MaxAge:           cfg.Server.CORS.MaxAge,
	}))
	e.Use(middleware.SecureWithConfig(middleware.SecureConfig{
		XSSProtection:         "1; mode=block",
		ContentTypeNosniff:    "nosniff",
		XFrameOptions:         "SAMEORIGIN",
		HSTSMaxAge:            31536000,
		HSTSExcludeSubdomains: false,
	}))

	// Swagger documentation in development
	if cfg.Swagger.Enabled {
		e.GET("/swagger/*", echoSwagger.WrapHandler)
	}

	return e
}

func initDependencies(db *gorm.DB, cfg *config.Config, log *zap.SugaredLogger) (repository.Factory, service.Factory) {
	repoFactory := infrarepo.NewFactory(db, log)
	serviceFactory := infraservice.NewFactory(repoFactory, cfg, log)
	return repoFactory, serviceFactory
}

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	logger, err := initLogger(&cfg.Logger)
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()
	sugar := logger.Sugar()

	// Initialize database
	db, err := database.NewPostgresDB(sugar, &cfg.Database)
	if err != nil {
		sugar.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Initialize Echo
	e := setupEcho(cfg, sugar)

	// Initialize dependencies
	repoFactory, serviceFactory := initDependencies(db.GormDB(), cfg, sugar)
	auth := serviceFactory.NewAuthService()
	authMiddleware := authMiddleware.NewAuthMiddleware(auth, sugar)

	// Initialize handlers
	handler.NewHealthHandler(e, sugar, repoFactory, serviceFactory)
	handler.NewAuthHandler(e, sugar, auth)
	handler.NewCategoryHandler(e, sugar, serviceFactory.NewCategoryService(), authMiddleware)
	handler.NewTransactionHandler(e, sugar, serviceFactory.NewTransactionService(), authMiddleware)
	handler.NewMonobankHandler(e, sugar, serviceFactory.NewMonobankService(), authMiddleware)

	// Start server
	go func() {
		if err := e.Start(":" + cfg.Server.Port); err != nil && err != http.ErrServerClosed {
			sugar.Fatalf("Failed to start server: %v", err)
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
		sugar.Fatalf("Failed to shutdown server: %v", err)
	}
}
