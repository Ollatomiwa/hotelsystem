package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"
)

type Config struct {
	Host string
	Port string
	User string
	Password string
	DBName string
	SSLMode string 
}

func NewPostgresConn(cfg Config) (*sql.DB, error) {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName,cfg.SSLMode)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open db connection: %w", err)
	}

	//configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 *time.Minute)

	//verifying connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to pinf db: %w", err)
	}
	log.Println("Succesfully connected to postgres db")
	return db, nil 
}

func InitializeSchema(db *sql.DB) error {
    queries := []string{
        `CREATE TABLE IF NOT EXISTS rooms (
            id TEXT PRIMARY KEY,
            room_number TEXT UNIQUE NOT NULL,
            room_type TEXT NOT NULL CHECK (room_type IN ('single', 'double', 'suite', 'deluxe')),
            price_per_night DECIMAL(10,2) NOT NULL,
            max_guests INTEGER NOT NULL CHECK (max_guests > 0),
            available BOOLEAN DEFAULT TRUE,
            description TEXT,
            created_at TIMESTAMPTZ DEFAULT NOW(),
            updated_at TIMESTAMPTZ DEFAULT NOW()
        )`,
        
        `CREATE TABLE IF NOT EXISTS bookings (
            id TEXT PRIMARY KEY,
            user_id TEXT NOT NULL,
            room_id TEXT NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
            room_type TEXT NOT NULL CHECK (room_type IN ('single', 'double', 'suite', 'deluxe')),
            check_in TIMESTAMPTZ NOT NULL,
            check_out TIMESTAMPTZ NOT NULL,
            guests INTEGER NOT NULL CHECK (guests > 0),
            total_amount DECIMAL(10,2) NOT NULL,
            status TEXT NOT NULL CHECK (status IN ('pending', 'confirmed', 'cancelled', 'completed')) DEFAULT 'pending',
            created_at TIMESTAMPTZ DEFAULT NOW(),
            updated_at TIMESTAMPTZ DEFAULT NOW(),
            CONSTRAINT valid_dates CHECK (check_out > check_in)
        )`,
        
        `CREATE INDEX IF NOT EXISTS idx_bookings_dates ON bookings (check_in, check_out)`,
        `CREATE INDEX IF NOT EXISTS idx_bookings_room_dates ON bookings (room_id, check_in, check_out)`,
        `CREATE INDEX IF NOT EXISTS idx_bookings_user ON bookings (user_id)`,
        `CREATE INDEX IF NOT EXISTS idx_rooms_type_available ON rooms (room_type, available)`,
    }

	for _, query := range queries {
		_, err := db.Exec(query)
		if err != nil {
			return fmt.Errorf("failed to execute query: %s, error: %w", query, err)
		}

		log.Println("Database schema initialized successfully")
	}
	return nil 
}