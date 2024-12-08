package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config represents the application's configuration
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Logger   LoggerConfig   `mapstructure:"logger"`
	Swagger  SwaggerConfig  `mapstructure:"swagger"`
	Metrics  MetricsConfig  `mapstructure:"metrics"`
	Features FeaturesConfig `mapstructure:"features"`
	Auth     AuthConfig     `mapstructure:"auth"`
	Security SecurityConfig `mapstructure:"security"`
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	Port    string     `mapstructure:"port"`
	Env     string     `mapstructure:"env"`
	Timeout int        `mapstructure:"timeout"`
	CORS    CORSConfig `mapstructure:"cors"`
}

// CORSConfig holds CORS-related configuration
type CORSConfig struct {
	AllowedOrigins   []string `mapstructure:"allowed_origins"`
	AllowedMethods   []string `mapstructure:"allowed_methods"`
	AllowedHeaders   []string `mapstructure:"allowed_headers"`
	AllowCredentials bool     `mapstructure:"allow_credentials"`
	MaxAge           int      `mapstructure:"max_age"`
}

// DatabaseConfig holds database-related configuration
type DatabaseConfig struct {
	Host            string        `mapstructure:"host"`
	Port            string        `mapstructure:"port"`
	User            string        `mapstructure:"user"`
	Password        string        `mapstructure:"password"`
	Name            string        `mapstructure:"name"`
	SSLMode         string        `mapstructure:"ssl_mode"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
}

// LoggerConfig holds logging-related configuration
type LoggerConfig struct {
	Level            string   `mapstructure:"level"`
	Encoding         string   `mapstructure:"encoding"`
	OutputPaths      []string `mapstructure:"output_paths"`
	ErrorOutputPaths []string `mapstructure:"error_output_paths"`
}

// SwaggerConfig holds Swagger documentation configuration
type SwaggerConfig struct {
	Enabled bool     `mapstructure:"enabled"`
	Host    string   `mapstructure:"host"`
	Schemes []string `mapstructure:"schemes"`
}

// MetricsConfig holds metrics-related configuration
type MetricsConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Path    string `mapstructure:"path"`
}

// FeaturesConfig holds feature flag configuration
type FeaturesConfig struct {
	MonobankIntegration bool `mapstructure:"monobank_integration"`
}

// AuthConfig holds authentication-related configuration
type AuthConfig struct {
	AccessTokenSecret  string `mapstructure:"access_token_secret"`
	RefreshTokenSecret string `mapstructure:"refresh_token_secret"`
	AccessTokenTTL     string `mapstructure:"access_token_ttl"`
	RefreshTokenTTL    string `mapstructure:"refresh_token_ttl"`
}

// SecurityConfig holds security-related configuration
type SecurityConfig struct {
	JWT JWTConfig `mapstructure:"jwt"`
}

// JWTConfig holds JWT-specific configuration
type JWTConfig struct {
	Secret                 string        `mapstructure:"secret"`
	AccessTokenExpiration  time.Duration `mapstructure:"access_token_expiration"`
	RefreshTokenExpiration time.Duration `mapstructure:"refresh_token_expiration"`
	Issuer                 string        `mapstructure:"issuer"`
	Audience               string        `mapstructure:"audience"`
}

// Load loads the configuration from files and environment variables
func Load() (*Config, error) {
	v := viper.New()

	env := os.Getenv("APP_ENV")
	if env == "" {
		env = os.Getenv("ENV") // Fallback to ENV for backward compatibility
	}
	if env == "" {
		env = "development"
	}

	configPath := "config"
	if customPath := os.Getenv("CONFIG_PATH"); customPath != "" {
		configPath = customPath
	}

	// Add config paths including subdirectories
	searchPaths := []string{
		configPath,
		configPath + "/env",
	}

	for _, path := range searchPaths {
		v.AddConfigPath(path)
	}
	v.SetConfigName(fmt.Sprintf("config.%s", env))
	v.SetConfigType("yaml")

	// Set environment variable prefix
	v.SetEnvPrefix("CASHONE")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Set default values
	setDefaults(v)

	// Read the config file
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
		return nil, fmt.Errorf("config file not found in paths: %v", searchPaths)
	}

	// Bind environment variables explicitly
	v.BindEnv("security.jwt.secret", "CASHONE_JWT_SECRET")
	v.BindEnv("database.name", "CASHONE_DATABASE_NAME")
	v.BindEnv("database.user", "CASHONE_DATABASE_USER")
	v.BindEnv("database.password", "CASHONE_DATABASE_PASSWORD")
	v.BindEnv("server.port", "CASHONE_SERVER_PORT")

	// If JWT secret is set in environment, override the default
	if jwtSecret := os.Getenv("CASHONE_JWT_SECRET"); jwtSecret != "" {
		v.Set("security.jwt.secret", jwtSecret)
	}

	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Set environment-specific values
	config.Server.Env = env
	if env == "production" {
		config.Swagger.Enabled = false
	}

	return &config, nil
}

func setDefaults(v *viper.Viper) {
	// Server defaults
	v.SetDefault("server.port", "3000")
	v.SetDefault("server.env", "development")
	v.SetDefault("server.timeout", 30)
	v.SetDefault("server.cors.allowed_origins", []string{"*"})
	v.SetDefault("server.cors.allowed_methods", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})
	v.SetDefault("server.cors.allowed_headers", []string{"*"})
	v.SetDefault("server.cors.allow_credentials", true)
	v.SetDefault("server.cors.max_age", 300)

	// Database defaults
	v.SetDefault("database.host", "localhost")
	v.SetDefault("database.port", "5432")
	v.SetDefault("database.user", "postgres")
	v.SetDefault("database.password", "postgres")
	v.SetDefault("database.name", "cashone")
	v.SetDefault("database.ssl_mode", "disable")
	v.SetDefault("database.max_open_conns", 25)
	v.SetDefault("database.max_idle_conns", 25)
	v.SetDefault("database.conn_max_lifetime", 300)

	// Logger defaults
	v.SetDefault("logger.level", "info")
	v.SetDefault("logger.encoding", "json")
	v.SetDefault("logger.output_paths", []string{"stdout"})
	v.SetDefault("logger.error_output_paths", []string{"stderr"})

	// Swagger defaults
	v.SetDefault("swagger.enabled", true)
	v.SetDefault("swagger.host", "localhost:3000")
	v.SetDefault("swagger.schemes", []string{"http"})

	// Metrics defaults
	v.SetDefault("metrics.enabled", true)
	v.SetDefault("metrics.path", "/metrics")

	// Features defaults
	v.SetDefault("features.monobank_integration", true)

	// Auth defaults
	v.SetDefault("auth.access_token_ttl", "15m")
	v.SetDefault("auth.refresh_token_ttl", "7d")

	// Security defaults
	v.SetDefault("security.jwt.secret", "your-jwt-secret-key")
	v.SetDefault("security.jwt.access_token_expiration", 15*time.Minute)
	v.SetDefault("security.jwt.refresh_token_expiration", 7*24*time.Hour)
	v.SetDefault("security.jwt.issuer", "cashone")
	v.SetDefault("security.jwt.audience", "cashone-users")
}
