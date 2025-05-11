package models

import (
	"database/sql"
	"fmt"
)

func Migrate(db *sql.DB) error {
	_, err := db.Exec(`
	CREATE TABLE IF NOT EXISTS users (
		id TEXT PRIMARY KEY,
		login TEXT UNIQUE NOT NULL,
		password TEXT NOT NULL
	);

	CREATE TABLE IF NOT EXISTS expressions (
		id TEXT PRIMARY KEY,
		expression TEXT NOT NULL,
		status TEXT NOT NULL,
		result REAL,
		user_id TEXT NOT NULL,
		FOREIGN KEY(user_id) REFERENCES users(id)
	);
	`)
	if err != nil {
		return fmt.Errorf("migrate: %w", err)
	}
	return nil
}
