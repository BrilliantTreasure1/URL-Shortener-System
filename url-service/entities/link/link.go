package link

import (
	"encoding/json"
	"time"
	"url-shortener/application/link/dto"
)

type Link struct {
	id *int
	userID int
	originalURL string
	shortCode   string
	createdAt   time.Time
	expiresAt   *time.Time
	isActive    bool
}

func NewLink(
	id *int,
	userID int,
	originalURL string,
	shortCode   string,
) (*Link , error) {
	
	return &Link{
		id:          id,
		userID:      userID,
		originalURL: originalURL,
		shortCode:   shortCode,
		createdAt:   time.Now(),
		isActive:    true,
	} , nil
}

func NewLinkFromDatabase(
	id *int,
	userID int,
	originalURL string,
	shortCode string,
	createdAt time.Time,
	expiresAt *time.Time,
	isActive bool,
) (*Link, error) {

	return &Link{
		id:          id,
		userID:      userID,
		originalURL: originalURL,
		shortCode:   shortCode,
		createdAt:   createdAt,
		expiresAt:   expiresAt,
		isActive:    isActive,
	}, nil
}

func (l *Link) Id() *int {
	return l.id
}

func (l *Link) UserID() int {
	return l.userID
}

func (l *Link) OriginalURL() string {
	return l.originalURL
}

func (l *Link) ShortCode() string {
	return l.shortCode
}

func (l *Link) CreatedAt() time.Time {
	return l.createdAt
}

func (l *Link) ExpiresAt() *time.Time {
	return l.expiresAt
}

func (l *Link) IsActive() bool {
	return l.isActive
}

func (l *Link) Disable() {
	l.isActive = false
}

func (l *Link) IsExpired(now time.Time) bool {
	if l.expiresAt == nil {
		return false
	}

	return now.After(*l.expiresAt)
}


func (l *Link) IsAvailable(now time.Time) bool {
	if !l.isActive {
		return false
	}

	if l.IsExpired(now) {
		return false
	}

	return true
}

func (l *Link) ToResponse() dto.LinkResponse {
	return dto.LinkResponse{
		ID:          *l.id,
		OriginalURL: l.originalURL,
		ShortCode:   l.shortCode,
		CreatedAt:   l.createdAt.Format(time.RFC3339),
		IsActive:    l.isActive,
	}
}

type linkJSON struct {
	ID          *int       `json:"id"`
	UserID      int        `json:"user_id"`
	OriginalURL string     `json:"original_url"`
	ShortCode   string     `json:"short_code"`
	CreatedAt   time.Time  `json:"created_at"`
	ExpiresAt   *time.Time `json:"expires_at"`
	IsActive    bool       `json:"is_active"`
}

func (l *Link) MarshalJSON() ([]byte, error) {
	return json.Marshal(linkJSON{
		ID:          l.id,
		UserID:      l.userID,
		OriginalURL: l.originalURL,
		ShortCode:   l.shortCode,
		CreatedAt:   l.createdAt,
		ExpiresAt:   l.expiresAt,
		IsActive:    l.isActive,
	})
}

func (l *Link) UnmarshalJSON(data []byte) error {
	var rec linkJSON
	if err := json.Unmarshal(data, &rec); err != nil {
		return err
	}

	l.id = rec.ID
	l.userID = rec.UserID
	l.originalURL = rec.OriginalURL
	l.shortCode = rec.ShortCode
	l.createdAt = rec.CreatedAt
	l.expiresAt = rec.ExpiresAt
	l.isActive = rec.IsActive

	return nil
}