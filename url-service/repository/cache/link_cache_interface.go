package cache

import (
	"errors"

	entities "url-shortener/entities/link"
)

var ErrCacheUnavailable = errors.New("cache is unavailable")

type LinkCache interface {
	Get(shortCode string) (*entities.Link, bool, error)
	Set(link *entities.Link) error
	Delete(shortCode string) error
}