package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
)

type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

func NewPostgresConn(cfg Config) (*sql.DB, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Verify connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("Successfully connected to PostgreSQL database")
	return db, nil
}

// InitializeSchema creates the necessary tables if they don't exist
func InitializeSchema(db *sql.DB) error {
	queries := []string{
		// Users table
		`CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			email TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			first_name TEXT NOT NULL,
			last_name TEXT NOT NULL,
			phone TEXT,
			role TEXT NOT NULL CHECK (role IN ('customer', 'admin')) DEFAULT 'customer'
		)`,

		// Index for faster email lookups
		`CREATE INDEX IF NOT EXISTS idx_users_email ON users(email)`,

		// Insert sample admin user (optional)
		`INSERT INTO users (id, email, password_hash, first_name, last_name, role) 
		VALUES (
			'admin-123',
			'admin@hotel.com',
			'$2a$10$ExampleHashedPasswordForAdmin123',
			'System',
			'Administrator',
			'admin'
		) ON CONFLICT (id) DO NOTHING`,
	}

	for _, query := range queries {
		_, err := db.Exec(query)
		if err != nil {
			return fmt.Errorf("failed to execute query: %s, error: %w", query, err)
		}
	}

	log.Println(" Database schema initialized successfully")
	return nil
}
