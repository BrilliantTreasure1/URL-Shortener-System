package database

import (
	"database/sql"
)

func Migrate(db *sql.DB) error {

	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			username VARCHAR(255) NOT NULL,
			email VARCHAR(255) NOT NULL UNIQUE,
			phonenumber VARCHAR(20) NOT NULL,
			password VARCHAR(255) NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,

		`CREATE SEQUENCE IF NOT EXISTS link_code_seq`,

		`CREATE TABLE IF NOT EXISTS links (
			id SERIAL PRIMARY KEY,
			user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			original_url TEXT NOT NULL,
			short_code VARCHAR(64) NOT NULL UNIQUE,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			expires_at TIMESTAMPTZ,
			is_active BOOLEAN NOT NULL DEFAULT TRUE
		)`,

		`CREATE INDEX IF NOT EXISTS idx_links_user_id ON links(user_id)`,
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return err
		}
	}

	return nil
}