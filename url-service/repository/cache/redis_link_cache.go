package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisLinkCache struct {
	client *redis.Client
	ttl    time.Duration
}

const cacheKeyPrefix = "link:"

func NewRedisLinkCache(client *redis.Client, ttl time.Duration) *RedisLinkCache {
	return &RedisLinkCache{
		client: client,
		ttl:    ttl,
	}
}

func (r *RedisLinkCache) Get(shortCode string) (string, bool, error) {
	if r.client == nil {
		return "", false, ErrCacheUnavailable
	}

	val, err := r.client.Get(context.Background(), cacheKeyPrefix+shortCode).Result()
	if err == redis.Nil {
		return "", false, nil
	}
	if err != nil {
		return "", false, ErrCacheUnavailable
	}

	return val, true, nil
}

func (r *RedisLinkCache) Set(shortCode, originalURL string) error {
	if r.client == nil {
		return ErrCacheUnavailable
	}

	if err := r.client.Set(context.Background(), cacheKeyPrefix+shortCode, originalURL, r.ttl).Err(); err != nil {
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