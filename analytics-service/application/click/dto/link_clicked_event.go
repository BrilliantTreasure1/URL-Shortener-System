package dto

import "time"

type LinkClickedEvent struct {
	EventID     string    `json:"event_id"`
	UserID      int       `json:"user_id"`
	ShortCode   string    `json:"short_code"`
	OriginalURL string    `json:"original_url"`
	ClickedAt   time.Time `json:"clicked_at"`
}