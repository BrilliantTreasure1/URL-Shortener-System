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