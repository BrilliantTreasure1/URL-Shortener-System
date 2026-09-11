package entities

import (
	"errors"
	"time"
)

type ClickEvent struct {
	eventID     string
	userID      int
	shortCode   string
	originalURL string
	clickedAt   time.Time
	consumedAt  time.Time
}

func NewClickEvent(
	eventID string,
	userID int,
	shortCode string,
	originalURL string,
	clickedAt time.Time,
) (*ClickEvent, error) {

	if eventID == "" {
		return nil, errors.New("event id cannot be empty")
	}

	if userID <= 0 {
		return nil, errors.New("invalid user id")
	}

	if shortCode == "" {
		return nil, errors.New("short code cannot be empty")
	}

	if originalURL == "" {
		return nil, errors.New("original url cannot be empty")
	}

	return &ClickEvent{
		eventID:     eventID,
		userID:      userID,
		shortCode:   shortCode,
		originalURL: originalURL,
		clickedAt:   clickedAt,
		consumedAt:  time.Now(),
	}, nil
}

func (c *ClickEvent) EventID() string {
	return c.eventID
}

func (c *ClickEvent) UserID() int {
	return c.userID
}

func (c *ClickEvent) ShortCode() string {
	return c.shortCode
}

func (c *ClickEvent) OriginalURL() string {
	return c.originalURL
}

func (c *ClickEvent) ClickedAt() time.Time {
	return c.clickedAt
}

func (c *ClickEvent) ConsumedAt() time.Time {
	return c.consumedAt
}