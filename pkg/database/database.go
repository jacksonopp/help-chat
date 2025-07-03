package database

import (
	"database/sql"
	"fmt"
	"time"

	"dev.azure.com/clearpointhealth/ClearQuoteV3/_git/helpchat/internal/config"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	_ "modernc.org/sqlite"
)

// Database represents the database connection
type Database struct {
	DB *gorm.DB
}

// NewDatabase creates a new database connection
func NewDatabase(cfg *config.Config) (*Database, error) {
	// Use the pure Go SQLite driver
	db, err := sql.Open("sqlite", cfg.Database.FilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Create GORM DB with the pure Go driver
	gormDB, err := gorm.Open(sqlite.New(sqlite.Config{
		Conn: db,
	}), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool with more conservative settings
	sqlDB, err := gormDB.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Use more conservative connection pool settings for SQLite
	sqlDB.SetMaxIdleConns(1)
	sqlDB.SetMaxOpenConns(1)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Configure SQLite pragmas for better performance and memory management
	if _, err := sqlDB.Exec("PRAGMA journal_mode=WAL"); err != nil {
		return nil, fmt.Errorf("failed to set journal mode: %w", err)
	}
	if _, err := sqlDB.Exec("PRAGMA synchronous=NORMAL"); err != nil {
		return nil, fmt.Errorf("failed to set synchronous mode: %w", err)
	}
	if _, err := sqlDB.Exec("PRAGMA cache_size=1000"); err != nil {
		return nil, fmt.Errorf("failed to set cache size: %w", err)
	}
	if _, err := sqlDB.Exec("PRAGMA temp_store=MEMORY"); err != nil {
		return nil, fmt.Errorf("failed to set temp store: %w", err)
	}

	// Test connection
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &Database{DB: gormDB}, nil
}

// Ping checks if the database is reachable
func (d *Database) Ping() error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}
	return sqlDB.Ping()
}

// Close closes the database connection
func (d *Database) Close() error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}
	return sqlDB.Close()
}
