package cache

import (
	"errors"
)

var ErrCacheUnavailable = errors.New("cache is unavailable")

type LinkCache interface {
	Get(shortCode string) (originalURL string, found bool, err error)
	Set(shortCode, originalURL string) error
	Delete(shortCode string) error
}