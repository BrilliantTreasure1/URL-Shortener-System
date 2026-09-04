package dto

type LinkResponse struct {
	ID          int    `json:"id"`
	OriginalURL string `json:"original_url"`
	ShortCode   string `json:"short_code"`
	CreatedAt   string `json:"created_at"`
	IsActive    bool   `json:"is_active"`
}
