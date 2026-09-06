package link

import (
	"errors"
	"time"

	entities "url-shortener/entities/link"
	linkRepo "url-shortener/repository/link"
	linkCache "url-shortener/repository/cache"
)

type ResolveShortLinkUseCase struct {
	linkRepo linkRepo.LinkRepository
	cache    linkCache.LinkCache
}

func NewResolveShortLinkUseCase(linkRepo linkRepo.LinkRepository, cache linkCache.LinkCache) *ResolveShortLinkUseCase {
	return &ResolveShortLinkUseCase{
		linkRepo: linkRepo,
		cache:    cache,
	}
}

func (r *ResolveShortLinkUseCase) ResolveShortLink(shortCode string) (*entities.Link, error) {

	if shortCode == "" {
		return nil, errors.New("short code cannot be empty")
	}

	if r.cache != nil {
		originalURL, found, err := r.cache.Get(shortCode)
		if err == nil && found {
			resolvedLink, _ := entities.NewLinkFromDatabase(nil, 0, originalURL, shortCode, time.Time{}, nil, true)
			return resolvedLink, nil
		}
	}

	link, err := r.linkRepo.FindByShortCode(shortCode)
	if err != nil {
		return nil, err
	}

	if link == nil {
		return nil, errors.New("link not found")
	}

	if !link.IsAvailable(time.Now()) {
		return nil, errors.New("link is not available")
	}

	if r.cache != nil {
		_ = r.cache.Set(shortCode, link.OriginalURL())
	}

	return link, nil
}