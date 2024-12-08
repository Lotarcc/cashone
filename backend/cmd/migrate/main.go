package main

import (
	"cashone/infrastructure/database"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// loadEnv loads environment variables from .env file
func loadEnv() {
	// Try to find .env file in parent directories
	dir := "../.."
	for i := 0; i < 3; i++ {
		envFile := filepath.Join(dir, ".env")
		if _, err := os.Stat(envFile); err == nil {
			content, err := os.ReadFile(envFile)
			if err != nil {
				log.Printf("Error reading .env file: %v", err)
				return
			}

			// Parse and set environment variables
			for _, line := range strings.Split(string(content), "\n") {
				line = strings.TrimSpace(line)
				if line == "" || strings.HasPrefix(line, "#") {
					continue
				}

				parts := strings.SplitN(line, "=", 2)
				if len(parts) != 2 {
					continue
				}

				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				value = strings.Trim(value, `"'`) // Remove quotes if present
				os.Setenv(key, value)
			}
			break
		}
		dir = filepath.Join(dir, "..")
	}
}

func main() {
	// Load environment variables
	loadEnv()

	// Parse command line arguments
	command := flag.String("command", "", "Migration command (up/down/status)")
	flag.Parse()

	if *command == "" {
		fmt.Println("Usage: migrate -command [up|down|status]")
		os.Exit(1)
	}

	// Get the absolute path to the config directory
	execPath, err := os.Executable()
	if err != nil {
		log.Fatalf("Error getting executable path: %v", err)
	}
	configPath := filepath.Join(filepath.Dir(execPath), "..", "..", "config")

	// Load configuration
	viper.SetConfigName("config.development")
	viper.SetConfigType("yaml")

	// Check CONFIG_PATH environment variable first
	if envConfigPath := os.Getenv("CONFIG_PATH"); envConfigPath != "" {
		viper.AddConfigPath(envConfigPath)
	}

	// Add standard config paths
	viper.AddConfigPath(configPath)
	viper.AddConfigPath(filepath.Join(configPath, "env")) // Add env subdirectory
	viper.AddConfigPath("../../config")                   // Fallback for development
	viper.AddConfigPath("../../config/env")               // Fallback for development env subdirectory

	// Enable environment variable replacement
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}

	// Database configuration
	dbConfig := viper.GetStringMapString("database")
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbConfig["host"],
		dbConfig["port"],
		os.Getenv("CASHONE_DATABASE_USER"),
		os.Getenv("CASHONE_DATABASE_PASSWORD"),
		os.Getenv("CASHONE_DATABASE_NAME"),
	)

	// Connect to database
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Create migration manager
	migrationManager := database.NewMigrationManager(db)

	// Execute command
	var cmdErr error
	switch *command {
	case "up":
		cmdErr = migrationManager.MigrateUp()
	case "down":
		cmdErr = migrationManager.MigrateDown()
	case "status":
		cmdErr = migrationManager.Status()
	default:
		log.Fatalf("Invalid command: %s", *command)
	}

	if cmdErr != nil {
		log.Fatalf("Migration error: %v", cmdErr)
	}
}
