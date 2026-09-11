package clickevent

import (
	"database/sql"

	entities "analytics-service/entities"
)

type ClickEventRepositoryPostgresql struct {
	db *sql.DB
}

func NewClickEventRepositoryPostgresql(db *sql.DB) *ClickEventRepositoryPostgresql {
	return &ClickEventRepositoryPostgresql{
		db: db,
	}
}

func (r *ClickEventRepositoryPostgresql) Save(clickEvent *entities.ClickEvent) (bool, error) {

	query := `
		INSERT INTO click_events (
			event_id,
			user_id,
			short_code,
			original_url,
			clicked_at,
			consumed_at
		)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (event_id) DO NOTHING
	`

	result, err := r.db.Exec(
		query,
		clickEvent.EventID(),
		clickEvent.UserID(),
		clickEvent.ShortCode(),
		clickEvent.OriginalURL(),
		clickEvent.ClickedAt(),
		clickEvent.ConsumedAt(),
	)
	if err != nil {
		return false, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}

	return rowsAffected > 0, nil
}

var _ ClickEventRepository = (*ClickEventRepositoryPostgresql)(nil)