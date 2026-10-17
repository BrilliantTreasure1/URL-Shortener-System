package link

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"

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

func (r *ResolveShortLinkUseCase) ResolveShortLink(ctx context.Context, shortCode string) (*entities.Link, error) {

	if shortCode == "" {
		return nil, errors.New("short code cannot be empty")
	}

	tracer := otel.Tracer("resolve")

	ucctx, span := tracer.Start(ctx, "usecase.resolve")
	defer span.End()

	span.SetAttributes(attribute.String("short_code", shortCode))

	if r.cache != nil {
		_, cacheSpan := tracer.Start(ucctx, "cache.get")
		cachedLink, found, err := r.cache.Get(shortCode)
		cacheSpan.SetAttributes(attribute.Bool("cache.hit", err == nil && found))
		cacheSpan.End()

		if err == nil && found {
			if cachedLink.IsAvailable(time.Now()) {
				r.publishClickEvent(ucctx, cachedLink)
				return cachedLink, nil
			}
		}
	}

	_, dbSpan := tracer.Start(ucctx, "repo.find-by-short-code")
	link, err := r.linkRepo.FindByShortCode(shortCode)
	dbSpan.End()

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
		_, setSpan := tracer.Start(ucctx, "cache.set")
		_ = r.cache.Set(link)
		setSpan.End()
	}

	r.publishClickEvent(ucctx, link)

	return link, nil
}

func (r *ResolveShortLinkUseCase) publishClickEvent(ctx context.Context, link *entities.Link) {
	if r.queue == nil {
		return
	}

	_, span := otel.Tracer("resolve").Start(ctx, "queue.publish")
	defer span.End()

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