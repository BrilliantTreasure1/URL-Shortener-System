package link

import entities "url-shortener/entities/link"

type LinkRepository interface {
	Create(link *entities.Link) (*entities.Link, error)
}