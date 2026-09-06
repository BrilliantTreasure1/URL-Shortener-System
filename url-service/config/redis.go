package config

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

const defaultCacheTTL = 3600

func NewCacheTTL() time.Duration {
	seconds, err := strconv.Atoi(getEnv("CACHETTL", strconv.Itoa(defaultCacheTTL)))
	if err != nil || seconds <= 0 {
		return defaultCacheTTL * time.Second
	}

	return time.Duration(seconds) * time.Second
}

func NewRedisClient() (*redis.Client, error) {

	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", getEnv("REDISHOST", "localhost"), getEnv("REDISPORT", "6379")),
		Password: getEnv("REDISPASSWORD", "redis"),
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		client.Close()
		return nil, err
	}

	return client, nil
}