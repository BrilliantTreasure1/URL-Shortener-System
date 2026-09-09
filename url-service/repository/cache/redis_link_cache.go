package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	entities "url-shortener/entities/link"
)

type RedisLinkCache struct {
	client *redis.Client
	ttl    time.Duration
}

type cachedLink struct {
	UserID      int        `json:"user_id"`
	ShortCode   string     `json:"short_code"`
	OriginalURL string     `json:"original_url"`
	ExpiresAt   *time.Time `json:"expires_at"`
	IsActive    bool       `json:"is_active"`
}

const cacheKeyPrefix = "link:"

func NewRedisLinkCache(client *redis.Client, ttl time.Duration) *RedisLinkCache {
	return &RedisLinkCache{
		client: client,
		ttl:    ttl,
	}
}

func (r *RedisLinkCache) Get(shortCode string) (*entities.Link, bool, error) {
	if r.client == nil {
		return nil, false, ErrCacheUnavailable
	}

	val, err := r.client.Get(context.Background(), cacheKeyPrefix+shortCode).Result()
	if err == redis.Nil {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, ErrCacheUnavailable
	}

	var rec cachedLink
	if err := json.Unmarshal([]byte(val), &rec); err != nil {
		return nil, false, nil
	}

	link, _ := entities.NewLinkWithState(
		nil,
		rec.UserID,
		rec.OriginalURL,
		rec.ShortCode,
		time.Time{},
		rec.ExpiresAt,
		rec.IsActive,
	)

	return link, true, nil
}

func (r *RedisLinkCache) Set(link *entities.Link) error {
	if r.client == nil {
		return ErrCacheUnavailable
	}

	rec := cachedLink{
		UserID:      link.UserID(),
		ShortCode:   link.ShortCode(),
		OriginalURL: link.OriginalURL(),
		ExpiresAt:   link.ExpiresAt(),
		IsActive:    link.IsActive(),
	}

	val, err := json.Marshal(rec)
	if err != nil {
		return fmt.Errorf("cache marshal failed: %w", err)
	}

	if err := r.client.Set(context.Background(), cacheKeyPrefix+link.ShortCode(), val, r.ttl).Err(); err != nil {
		return fmt.Errorf("cache set failed: %w", err)
	}

	return nil
}

func (r *RedisLinkCache) Delete(shortCode string) error {
	if r.client == nil {
		return ErrCacheUnavailable
	}

	if err := r.client.Del(context.Background(), cacheKeyPrefix+shortCode).Err(); err != nil {
		return fmt.Errorf("cache delete failed: %w", err)
	}

	return nil
}