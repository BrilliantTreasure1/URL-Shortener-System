package link

import entities "url-shortener/entities/link"

type LinkRepository interface {
	Create(link *entities.Link) (*entities.Link, error)
	NextCodeValue() (int64, error)
	FindByShortCode(shortCode string) (*entities.Link , error)
	ListByUserID(userID int, offset int, limit int) ([]*entities.Link, int64, error)
}