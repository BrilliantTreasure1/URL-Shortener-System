package report

import "database/sql"

type ReportRepositoryPostgresql struct {
	db *sql.DB
}

func NewReportRepositoryPostgresql(db *sql.DB) *ReportRepositoryPostgresql {
	return &ReportRepositoryPostgresql{
		db: db,
	}
}

func (r *ReportRepositoryPostgresql) CountsByUserID(userID int) (*Overview, error) {

	query := `
		SELECT
			(SELECT COUNT(*) FROM links
				WHERE user_id = $1) AS total_links,
			(SELECT COUNT(*) FROM links
				WHERE user_id = $1
				  AND is_active = true
				  AND (expires_at IS NULL OR expires_at > NOW())) AS active_links,
			(SELECT COUNT(*) FROM click_events
				WHERE user_id = $1) AS total_clicks
	`

	var overview Overview

	err := r.db.QueryRow(query, userID).Scan(
		&overview.TotalLinks,
		&overview.ActiveLinks,
		&overview.TotalClicks,
	)
	if err != nil {
		return nil, err
	}

	return &overview, nil
}

var _ ReportRepository = (*ReportRepositoryPostgresql)(nil)