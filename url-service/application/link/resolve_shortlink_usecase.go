package link

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"

	linkDTO "url-shortener/application/link/dto"
	entities "url-shortener/entities/link"
	linkCache "url-shortener/repository/cache"
	linkRepo "url-shortener/repository/link"
	messagequeue "url-shortener/message-queue"
)

type ResolveShortLinkUseCase struct {
	linkRepo linkRepo.LinkRepository
	cache    linkCache.LinkCache
	queue    messagequeue.Rabbitmq
}

func NewResolveShortLinkUseCase(linkRepo linkRepo.LinkRepository, cache linkCache.LinkCache, queue messagequeue.Rabbitmq) *ResolveShortLinkUseCase {
	return &ResolveShortLinkUseCase{
		linkRepo: linkRepo,
		cache:    cache,
		queue:    queue,
	}
}

func (r *ResolveShortLinkUseCase) ResolveShortLink(shortCode string) (*entities.Link, error) {

	if shortCode == "" {
		return nil, errors.New("short code cannot be empty")
	}

	if r.cache != nil {
		cachedLink, found, err := r.cache.Get(shortCode)
		if err == nil && found {
			if cachedLink.IsAvailable(time.Now()) {
				r.publishClickEvent(cachedLink)
				return cachedLink, nil
			}
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
		_ = r.cache.Set(link)
	}

	r.publishClickEvent(link)

	return link, nil
}

func (r *ResolveShortLinkUseCase) publishClickEvent(link *entities.Link) {
	if r.queue == nil {
		return
	}

	payload, err := json.Marshal(linkDTO.LinkClickedEvent{
		EventID:     uuid.NewString(),
		UserID:      link.UserID(),
		ShortCode:   link.ShortCode(),
		OriginalURL: link.OriginalURL(),
		ClickedAt:   time.Now(),
	})
	if err != nil {
		return
	}

	_ = r.queue.Publish(messagequeue.RoutingKeyLinkClicked, payload)
}


