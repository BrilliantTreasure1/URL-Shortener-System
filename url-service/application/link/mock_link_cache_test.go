package link

import (
	entities "url-shortener/entities/link"
	messagequeue "url-shortener/message-queue"
	linkCache "url-shortener/repository/cache"
)

type mockLinkCache struct {
	hitLink   *entities.Link
	hitFound  bool
	hitErr    error
	calledGet bool

	setErr   error
	calledSet bool
	setLink  *entities.Link

	deleteErr    error
	calledDelete bool
	deleteCode   string
}

func (m *mockLinkCache) Get(shortCode string) (*entities.Link, bool, error) {
	m.calledGet = true
	return m.hitLink, m.hitFound, m.hitErr
}

func (m *mockLinkCache) Set(link *entities.Link) error {
	m.calledSet = true
	m.setLink = link
	return m.setErr
}

func (m *mockLinkCache) Delete(shortCode string) error {
	m.calledDelete = true
	m.deleteCode = shortCode
	return m.deleteErr
}

type mockQueue struct {
	publishErr error

	calledPublish bool
	routingKey    string
	payload       []byte
}

func (m *mockQueue) Publish(routingKey string, payload []byte) error {
	m.calledPublish = true
	m.routingKey = routingKey
	m.payload = payload
	return m.publishErr
}

var _ linkCache.LinkCache = (*mockLinkCache)(nil)
var _ messagequeue.Rabbitmq = (*mockQueue)(nil)