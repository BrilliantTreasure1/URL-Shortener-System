package link

import(
	"time"
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