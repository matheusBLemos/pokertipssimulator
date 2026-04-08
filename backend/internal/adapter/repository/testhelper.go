package repository

import (
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

func NewTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open in-memory sqlite: %v", err)
	}
	db.SetMaxOpenConns(1)

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS rooms (
			id         TEXT PRIMARY KEY,
			code       TEXT UNIQUE NOT NULL,
			mode       TEXT NOT NULL DEFAULT 'game',
			data       TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		CREATE TABLE IF NOT EXISTS settings (
			key   TEXT PRIMARY KEY,
			value TEXT NOT NULL
		);
	`)
	if err != nil {
		db.Close()
		t.Fatalf("migrate in-memory sqlite: %v", err)
	}

	t.Cleanup(func() { db.Close() })
	return db
}
