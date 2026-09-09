package link

import (
	"time"

	entities "url-shortener/entities/link"
)

func newLinkWithState(
	id int,
	userID int,
	originalURL string,
	shortCode string,
	isActive bool,
	expiresAt *time.Time,
) *entities.Link {
	link, err := entities.NewLinkWithState(
		&id,
		userID,
		originalURL,
		shortCode,
		time.Now(),
		expiresAt,
		isActive,
	)
	if err != nil {
		panic(err)
	}
	return link
}

func pastTime() *time.Time {
	t := time.Now().Add(-time.Hour)
	return &t
}
