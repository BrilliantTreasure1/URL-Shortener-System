package link

import (
	"database/sql"
	"time"

	entities "url-shortener/entities/link"
)

type LinkRepositoryPostgresql struct {
	db *sql.DB
}

func NewLinkRepositoryPostgresql(db *sql.DB) *LinkRepositoryPostgresql {
	return &LinkRepositoryPostgresql{
		db: db,
	}
}

func (r *LinkRepositoryPostgresql) Create(
	link *entities.Link,
) (*entities.Link, error) {

	query := `
		INSERT INTO links (
			user_id,
			original_url,
			short_code
		)
		VALUES ($1, $2, $3)
		RETURNING
			id,
			user_id,
			original_url,
			short_code,
			created_at,
			expires_at,
			is_active
	`

	var (
		id          int
		userID      int
		originalURL string
		shortCode   string
		createdAt   time.Time
		expiresAt   sql.NullTime
		isActive    bool
	)

	err := r.db.QueryRow(
		query,
		link.UserID(),
		link.OriginalURL(),
		link.ShortCode(),
	).Scan(
		&id,
		&userID,
		&originalURL,
		&shortCode,
		&createdAt,
		&expiresAt,
		&isActive,
	)

	if err != nil {
		return nil, err
	}

	var expiresAtPtr *time.Time
	if expiresAt.Valid {
		t := expiresAt.Time
		expiresAtPtr = &t
	}

	return entities.NewLinkFromDatabase(
		&id,
		userID,
		originalURL,
		shortCode,
		createdAt,
		expiresAtPtr,
		isActive,
	)
}

func (r *LinkRepositoryPostgresql) NextCodeValue() (int64, error) {

	var value int64

	err := r.db.QueryRow(`SELECT nextval('link_code_seq')`).Scan(&value)
	if err != nil {
		return 0, err
	}

	return value, nil
}

	
func (r *LinkRepositoryPostgresql) FindByShortCode(shortCode string) (*entities.Link, error) {

	query := `
		SELECT
			id,
			user_id,
			original_url,
			short_code,
			created_at,
			expires_at,
			is_active
		FROM links
		WHERE short_code = $1
	`

	var (
		id          int
		userID      int
		originalURL string
		shortCodeDB string
		createdAt   time.Time
		expiresAt   sql.NullTime
		isActive    bool
	)

	err := r.db.QueryRow(query, shortCode).Scan(
		&id,
		&userID,
		&originalURL,
		&shortCodeDB,
		&createdAt,
		&expiresAt,
		&isActive,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	var expiresAtPtr *time.Time
	if expiresAt.Valid {
		t := expiresAt.Time
		expiresAtPtr = &t
	}

	return entities.NewLinkFromDatabase(
		&id,
		userID,
		originalURL,
		shortCodeDB,
		createdAt,
		expiresAtPtr,
		isActive,
	)
}