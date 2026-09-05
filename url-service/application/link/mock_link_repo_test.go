package link

import (
	entities "url-shortener/entities/link"
	linkRepo "url-shortener/repository/link"
)

type mockLinkRepo struct {
	createLink *entities.Link
	createErr  error
	calledCreate bool

	nextValue int64
	nextErr   error
	calledNext bool

	findByShortCodeLink *entities.Link
	findByShortCodeErr  error
	calledFindByShortCode bool
}

func (m *mockLinkRepo) Create(l *entities.Link) (*entities.Link, error) {
	m.calledCreate = true
	return m.createLink, m.createErr
}

func (m *mockLinkRepo) NextCodeValue() (int64, error) {
	m.calledNext = true
	return m.nextValue, m.nextErr
}

func (m *mockLinkRepo) FindByShortCode(shortCode string) (*entities.Link, error) {
	m.calledFindByShortCode = true
	return m.findByShortCodeLink, m.findByShortCodeErr
}

func (m *mockLinkRepo) ListByUserID(userID int, offset int, limit int) ([]*entities.Link, int64, error) {
	return []*entities.Link{}, 0, nil
}

func (m *mockLinkRepo) DisableLink(userID int, shortCode string) (*entities.Link, error) {
	return nil, nil
}

var _ linkRepo.LinkRepository = (*mockLinkRepo)(nil)