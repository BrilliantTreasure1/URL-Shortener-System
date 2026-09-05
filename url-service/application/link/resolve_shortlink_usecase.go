package link

import (
	"errors"
	"time"

	entities "url-shortener/entities/link"
	linkRepo "url-shortener/repository/link"
)

type ResolveShortLinkUseCase struct {
	linkRepo linkRepo.LinkRepository
}

func NewResolveShortLinkUseCase(linkRepo linkRepo.LinkRepository) *ResolveShortLinkUseCase {
	return &ResolveShortLinkUseCase{
		linkRepo: linkRepo,
	}
}

func (r *ResolveShortLinkUseCase) ResolveShortLink(shortCode string) (*entities.Link, error) {

	if shortCode == "" {
		return nil, errors.New("short code cannot be empty")
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

	return link, nil
}