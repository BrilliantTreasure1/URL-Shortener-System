package cache

import (
	"errors"
	"testing"
	"time"
)

func TestAllowAttemptWhenEnabled(t *testing.T) {
	c := NewRedisLinkCache(nil, time.Second)

	if !c.allowAttempt() {
		t.Fatal("expected enabled cache to allow attempts")
	}
	if c.disabled {
		t.Fatal("expected cache to stay enabled")
	}
}

func TestTripBlocksAttemptsDuringCooldown(t *testing.T) {
	old := redisReconnectCooldown
	redisReconnectCooldown = 30 * time.Second
	defer func() { redisReconnectCooldown = old }()

	c := NewRedisLinkCache(nil, time.Second)
	c.trip()

	if c.allowAttempt() {
		t.Fatal("expected attempt to be blocked during cooldown")
	}
	if !c.disabled {
		t.Fatal("expected cache to stay disabled during cooldown")
	}
}

func TestHalfOpenProbeAfterCooldown(t *testing.T) {
	old := redisReconnectCooldown
	redisReconnectCooldown = time.Second
	defer func() { redisReconnectCooldown = old }()

	c := NewRedisLinkCache(nil, time.Second)
	c.trip()
	c.disabledAt = time.Now().Add(-2 * time.Second)

	if !c.allowAttempt() {
		t.Fatal("expected probe attempt after cooldown elapsed")
	}
	if c.disabled {
		t.Fatal("expected cache to re-enable for probe")
	}
}

func TestCircuitBreakerReTripsAfterProbe(t *testing.T) {
	old := redisReconnectCooldown
	redisReconnectCooldown = time.Second
	defer func() { redisReconnectCooldown = old }()

	c := NewRedisLinkCache(nil, time.Second)

	c.trip()
	c.disabledAt = time.Now().Add(-2 * time.Second)
	if !c.allowAttempt() {
		t.Fatal("expected probe attempt after cooldown elapsed")
	}

	c.trip()
	if c.allowAttempt() {
		t.Fatal("expected attempt blocked again after fresh trip")
	}
}

func TestRedisLinkCacheNilClient(t *testing.T) {
	c := NewRedisLinkCache(nil, time.Second)

	if _, _, err := c.Get("abc"); !errors.Is(err, ErrCacheUnavailable) {
		t.Fatalf("Get: expected ErrCacheUnavailable, got %v", err)
	}
	if err := c.Set(nil); !errors.Is(err, ErrCacheUnavailable) {
		t.Fatalf("Set: expected ErrCacheUnavailable, got %v", err)
	}
	if err := c.Delete("abc"); !errors.Is(err, ErrCacheUnavailable) {
		t.Fatalf("Delete: expected ErrCacheUnavailable, got %v", err)
	}
}