package db

import (
	"database/sql"
	"encoding/json"
	"os"
	"time"

	_ "github.com/lib/pq"

	"github.com/laura838/shld-i-buy-this/models"
)

const schema = `
CREATE TABLE IF NOT EXISTS user_profile (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL DEFAULT '',
    purchase_patterns TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);
ALTER TABLE user_profile ADD COLUMN IF NOT EXISTS name TEXT NOT NULL DEFAULT '';

CREATE TABLE IF NOT EXISTS sessions (
    id SERIAL PRIMARY KEY,
    product TEXT NOT NULL,
    product_name TEXT NOT NULL DEFAULT '',
    verdict TEXT NOT NULL,
    reasoning TEXT NOT NULL,
    alternatives JSONB NOT NULL DEFAULT '[]',
    created_at TIMESTAMP NOT NULL
);
ALTER TABLE sessions ADD COLUMN IF NOT EXISTS product_name TEXT NOT NULL DEFAULT '';
ALTER TABLE sessions ADD COLUMN IF NOT EXISTS alternatives JSONB NOT NULL DEFAULT '[]';
`

// InitDB opens a connection to PostgreSQL and runs schema migrations.
func InitDB() (*sql.DB, error) {
	connStr := os.Getenv("DATABASE_URL")
	conn, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	if _, err := conn.Exec(schema); err != nil {
		conn.Close()
		return nil, err
	}
	return conn, nil
}

// WriteProfile saves the user profile. Overwrites any existing profile.
func WriteProfile(db *sql.DB, profile models.UserProfile) error {
	now := time.Now()
	_, err := db.Exec(`
		INSERT INTO user_profile (id, name, purchase_patterns, created_at, updated_at)
		VALUES (1, $1, $2, $3, $4)
		ON CONFLICT(id) DO UPDATE SET
			name = excluded.name,
			purchase_patterns = excluded.purchase_patterns,
			updated_at = excluded.updated_at
	`, profile.Name, profile.PurchasePatterns, now, now)
	return err
}

// ReadProfile fetches the user profile. Returns nil, nil if no profile exists.
func ReadProfile(db *sql.DB) (*models.UserProfile, error) {
	var name, purchasePatterns string
	err := db.QueryRow("SELECT name, purchase_patterns FROM user_profile WHERE id = 1").Scan(&name, &purchasePatterns)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &models.UserProfile{Name: name, PurchasePatterns: purchasePatterns}, nil
}

// WriteSession persists a completed evaluation session.
func WriteSession(db *sql.DB, session models.Session) error {
	altsJSON, err := json.Marshal(session.Alternatives)
	if err != nil {
		return err
	}
	_, err = db.Exec(`
		INSERT INTO sessions (product, product_name, verdict, reasoning, alternatives, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, session.Product, session.ProductName, session.Verdict, session.Reasoning, altsJSON, session.CreatedAt)
	return err
}
