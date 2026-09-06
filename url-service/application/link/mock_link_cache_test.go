package link

import (
	linkCache "url-shortener/repository/cache"
)

type mockLinkCache struct {
	hitURL   string
	hitFound bool
	hitErr   error

	setErr    error
	calledSet bool
	setCode   string
	setURL    string

	deleteErr    error
	calledDelete bool
	deleteCode   string
}

func (m *mockLinkCache) Get(shortCode string) (string, bool, error) {
	return m.hitURL, m.hitFound, m.hitErr
}

func (m *mockLinkCache) Set(shortCode, originalURL string) error {
	m.calledSet = true
	m.setCode = shortCode
	m.setURL = originalURL
	return m.setErr
}

func (m *mockLinkCache) Delete(shortCode string) error {
	m.calledDelete = true
	m.deleteCode = shortCode
	return m.deleteErr
}

var _ linkCache.LinkCache = (*mockLinkCache)(nil)