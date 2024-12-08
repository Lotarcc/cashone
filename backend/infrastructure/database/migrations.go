package database

import (
	"cashone/domain/entity"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"gorm.io/gorm"
)

// MigrationManager handles database migrations
type MigrationManager struct {
	db *gorm.DB
}

// NewMigrationManager creates a new migration manager
func NewMigrationManager(db *gorm.DB) *MigrationManager {
	return &MigrationManager{db: db}
}

// MigrateUp runs all migrations
func (m *MigrationManager) MigrateUp() error {
	// Create migrations table if it doesn't exist
	err := m.db.AutoMigrate(&entity.Migration{})
	if err != nil {
		return fmt.Errorf("failed to create migrations table: %v", err)
	}

	// Get all SQL migration files
	files, err := m.getMigrationFiles()
	if err != nil {
		return fmt.Errorf("failed to get migration files: %v", err)
	}

	// Run each migration in transaction
	for _, file := range files {
		version := strings.Split(filepath.Base(file), "_")[0]

		// Check if migration was already applied
		var count int64
		m.db.Model(&entity.Migration{}).Where("version = ?", version).Count(&count)
		if count > 0 {
			continue
		}

		// Read migration file
		content, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %v", file, err)
		}

		// Begin transaction
		tx := m.db.Begin()

		// Execute migration
		if err := tx.Exec(string(content)).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to execute migration %s: %v", file, err)
		}

		// Record migration
		if err := tx.Create(&entity.Migration{Version: version}).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to record migration %s: %v", file, err)
		}

		// Commit transaction
		if err := tx.Commit().Error; err != nil {
			return fmt.Errorf("failed to commit migration %s: %v", file, err)
		}

		log.Printf("Applied migration: %s\n", version)
	}

	return nil
}

// MigrateDown rolls back all migrations
func (m *MigrationManager) MigrateDown() error {
	var migrations []entity.Migration
	if err := m.db.Order("version DESC").Find(&migrations).Error; err != nil {
		return fmt.Errorf("failed to get applied migrations: %v", err)
	}

	for _, migration := range migrations {
		// Find corresponding down migration file
		downFile := filepath.Join(m.getMigrationsDir(), fmt.Sprintf("%s_down.sql", migration.Version))
		if _, err := os.Stat(downFile); os.IsNotExist(err) {
			return fmt.Errorf("down migration file not found for version %s", migration.Version)
		}

		// Read down migration file
		content, err := os.ReadFile(downFile)
		if err != nil {
			return fmt.Errorf("failed to read down migration %s: %v", downFile, err)
		}

		// Begin transaction
		tx := m.db.Begin()

		// Execute down migration
		if err := tx.Exec(string(content)).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to execute down migration %s: %v", downFile, err)
		}

		// Remove migration record
		if err := tx.Delete(&migration).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to remove migration record %s: %v", migration.Version, err)
		}

		// Commit transaction
		if err := tx.Commit().Error; err != nil {
			return fmt.Errorf("failed to commit down migration %s: %v", downFile, err)
		}

		log.Printf("Rolled back migration: %s\n", migration.Version)
	}

	return nil
}

// Status prints the status of all migrations
func (m *MigrationManager) Status() error {
	var migrations []entity.Migration
	if err := m.db.Order("version ASC").Find(&migrations).Error; err != nil {
		return fmt.Errorf("failed to get migrations: %v", err)
	}

	files, err := m.getMigrationFiles()
	if err != nil {
		return fmt.Errorf("failed to get migration files: %v", err)
	}

	fmt.Println("Migration Status:")
	fmt.Println("================")

	appliedVersions := make(map[string]bool)
	for _, migration := range migrations {
		appliedVersions[migration.Version] = true
		fmt.Printf("[✓] %s (applied)\n", migration.Version)
	}

	for _, file := range files {
		version := strings.Split(filepath.Base(file), "_")[0]
		if !appliedVersions[version] {
			fmt.Printf("[ ] %s (pending)\n", version)
		}
	}

	return nil
}

func (m *MigrationManager) getMigrationsDir() string {
	// Try to find the db/migrations directory relative to the current working directory
	dir, err := os.Getwd()
	if err != nil {
		log.Printf("Warning: Could not get working directory: %v", err)
		return "db/migrations"
	}

	// Look for db/migrations directory
	migrationsPath := filepath.Join(dir, "db", "migrations")
	if _, err := os.Stat(migrationsPath); err == nil {
		return migrationsPath
	}

	// If not found, try parent directory
	parentDir := filepath.Dir(dir)
	migrationsPath = filepath.Join(parentDir, "db", "migrations")
	if _, err := os.Stat(migrationsPath); err == nil {
		return migrationsPath
	}

	// Default to db/migrations in current directory
	return "db/migrations"
}

func (m *MigrationManager) getMigrationFiles() ([]string, error) {
	migrationsDir := m.getMigrationsDir()
	var files []string
	err := filepath.Walk(migrationsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".sql") && !strings.HasSuffix(path, "_down.sql") {
			// Skip template files
			if !strings.HasPrefix(filepath.Base(path), "template") {
				files = append(files, path)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return files, nil
}
