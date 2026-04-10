package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

// Client represents the database connection
type Client struct {
	db *sql.DB
}

// NewSQLiteClient creates a new SQLite database connection
func NewSQLiteClient(dbPath string) (*Client, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	client := &Client{db: db}

	// Run migrations
	if err := client.runMigrations(); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Println("Database initialized successfully")
	return client, nil
}

// runMigrations executes all SQL migration files
func (c *Client) runMigrations() error {
	// Find migrations directory relative to executable or current directory
	migrationsPath := "./migrations"

	// Try relative path from current directory
	if _, err := os.Stat(migrationsPath); os.IsNotExist(err) {
		// Try relative from parent directory (for cmd/*)
		migrationsPath = "../migrations"
		if _, err := os.Stat(migrationsPath); os.IsNotExist(err) {
			// Try from grandparent
			migrationsPath = "../../migrations"
		}
	}

	log.Printf("Loading migrations from: %s", migrationsPath)

	// Read all migration files
	entries, err := os.ReadDir(migrationsPath)
	if err != nil {
		log.Printf("Warning: could not read migrations directory at %s: %v", migrationsPath, err)
		log.Println("Continuing without running migrations - schema might not exist!")
		return nil // Don't fail, just warn
	}

	// Sort files to ensure execution order
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}

		filePath := filepath.Join(migrationsPath, entry.Name())
		migrationSQL, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read migration %s: %w", entry.Name(), err)
		}

		// Execute migration
		if _, err := c.db.Exec(string(migrationSQL)); err != nil {
			return fmt.Errorf("failed to execute migration %s: %w", entry.Name(), err)
		}

		log.Printf("✓ Executed migration: %s", entry.Name())
	}

	return nil
}

// Close closes the database connection
func (c *Client) Close() error {
	return c.db.Close()
}

// GetDB returns the underlying sql.DB for advanced usage
func (c *Client) GetDB() *sql.DB {
	return c.db
}

// BeginTx starts a new database transaction
func (c *Client) BeginTx() (*sql.Tx, error) {
	return c.db.Begin()
}
