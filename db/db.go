package db

import (
	"database/sql"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/laura838/shld-i-buy-this/models"
)

const schema = `
CREATE TABLE IF NOT EXISTS user_profile (
    id INTEGER PRIMARY KEY,
    purchase_patterns TEXT NOT NULL,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL
);

CREATE TABLE IF NOT EXISTS sessions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    product TEXT NOT NULL,
    verdict TEXT NOT NULL,
    reasoning TEXT NOT NULL,
    created_at DATETIME NOT NULL
);
`

// InitDB opens (or creates) the SQLite database file and runs schema migrations.
func InitDB(path string) (*sql.DB, error) {
	conn, err := sql.Open("sqlite3", path)
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
		INSERT INTO user_profile (id, purchase_patterns, created_at, updated_at)
		VALUES (1, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			purchase_patterns = excluded.purchase_patterns,
			updated_at = excluded.updated_at
	`, profile.PurchasePatterns, now, now)
	return err
}

// ReadProfile fetches the user profile. Returns nil, nil if no profile exists.
func ReadProfile(db *sql.DB) (*models.UserProfile, error) {
	var purchasePatterns string
	err := db.QueryRow("SELECT purchase_patterns FROM user_profile WHERE id = 1").Scan(&purchasePatterns)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &models.UserProfile{PurchasePatterns: purchasePatterns}, nil
}

// WriteSession persists a completed evaluation session.
func WriteSession(db *sql.DB, session models.Session) error {
	_, err := db.Exec(`
		INSERT INTO sessions (product, verdict, reasoning, created_at)
		VALUES (?, ?, ?, ?)
	`, session.Product, session.Verdict, session.Reasoning, session.CreatedAt)
	return err
}
