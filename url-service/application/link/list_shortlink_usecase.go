package link

import (
	"errors"

	entities "url-shortener/entities/link"
	linkRepo "url-shortener/repository/link"
)

type ListShortLinkUseCase struct {
	linkRepo linkRepo.LinkRepository
}

func NewListShortLinkUseCase(linkRepo linkRepo.LinkRepository) *ListShortLinkUseCase {
	return &ListShortLinkUseCase{
		linkRepo: linkRepo,
	}
}

func (l *ListShortLinkUseCase) ListShortLink(userID int, offset int, limit int) ([]*entities.Link, int64, error) {

	if userID <= 0 {
		return nil, 0, errors.New("invalid user id")
	}

	if offset < 0 {
		return nil, 0, errors.New("offset cannot be negative")
	}

	if limit <= 0 {
		return nil, 0, errors.New("limit must be greater than zero")
	}

	links, total, err := l.linkRepo.FindByUserID(userID, offset, limit)
	if err != nil {
		return nil, 0, err
	}

	return links, total, nil
}