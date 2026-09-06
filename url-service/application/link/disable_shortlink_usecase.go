package link

import (
	"errors"

	entities "url-shortener/entities/link"
	linkRepo "url-shortener/repository/link"
	linkCache "url-shortener/repository/cache"
)

var ErrLinkAlreadyDisabled = errors.New("link is already disabled")

type DisableShortLinkUseCase struct {
	linkRepo linkRepo.LinkRepository
	cache    linkCache.LinkCache
}

func NewDisableShortLinkUseCase(linkRepo linkRepo.LinkRepository, cache linkCache.LinkCache) *DisableShortLinkUseCase {
	return &DisableShortLinkUseCase{
		linkRepo: linkRepo,
		cache:    cache,
	}
}

func (d *DisableShortLinkUseCase) DisableShortLink(userID int, shortCode string) (*entities.Link, error) {

	if userID <= 0 {
		return nil, errors.New("invalid user id")
	}

	if shortCode == "" {
		return nil, errors.New("short code cannot be empty")
	}

	link, err := d.linkRepo.DisableLink(userID, shortCode)
	if err != nil {
		if errors.Is(err, linkRepo.ErrLinkAlreadyDisabled) {
			return nil, ErrLinkAlreadyDisabled
		}
		return nil, err
	}

	if link == nil {
		return nil, errors.New("link not found")
	}

	if d.cache != nil {
		_ = d.cache.Delete(shortCode)
	}

	return link, nil
}