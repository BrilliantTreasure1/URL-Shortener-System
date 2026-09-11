package database

import "database/sql"

func Migrate(db *sql.DB) error {

	queries := []string{
		`CREATE TABLE IF NOT EXISTS click_events (
			event_id     VARCHAR(36) PRIMARY KEY,
			user_id      INTEGER     NOT NULL,
			short_code   VARCHAR(64) NOT NULL,
			original_url TEXT        NOT NULL,
			clicked_at   TIMESTAMPTZ NOT NULL,
			consumed_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,

		`CREATE INDEX IF NOT EXISTS idx_click_events_user_id ON click_events(user_id)`,

		`CREATE INDEX IF NOT EXISTS idx_click_events_short_code ON click_events(short_code)`,
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return err
		}
	}

	return nil
}