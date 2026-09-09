package link

import (
	"encoding/json"
	"errors"
	"testing"

	linkDTO "url-shortener/application/link/dto"
	messagequeue "url-shortener/message-queue"
	linkCache "url-shortener/repository/cache"
)

func TestResolveShortLink(t *testing.T) {
	repoErr := errors.New("repository failed")

	tests := []struct {
		name           string
		shortCode      string
		repo           *mockLinkRepo
		cache          *mockLinkCache
		queue          *mockQueue
		wantErr        string
		wantShortCode  string
		wantOrigURL    string
		wantUserID     int
		wantFindCalled bool
		wantSet        bool
		wantPublish    bool
	}{
		{
			name:      "finds link",
			shortCode: "abc123",
			repo: &mockLinkRepo{
				findByShortCodeLink: mustLink(
					10, 1, "https://example.com", "abc123",
				),
			},
			cache:          &mockLinkCache{},
			queue:          &mockQueue{},
			wantErr:        "",
			wantShortCode:  "abc123",
			wantOrigURL:    "https://example.com",
			wantUserID:     1,
			wantFindCalled: true,
			wantSet:        true,
			wantPublish:    true,
		},
		{
			name:      "serves from cache",
			shortCode: "abc123",
			repo:      &mockLinkRepo{},
			cache: &mockLinkCache{
				hitLink: mustLink(
					10, 1, "https://cached.example.com", "abc123",
				),
				hitFound: true,
			},
			queue:       &mockQueue{},
			wantErr:     "",
			wantShortCode: "abc123",
			wantOrigURL: "https://cached.example.com",
			wantUserID:  1,
			wantSet:     false,
			wantPublish: true,
		},
		{
			name:      "cached link unavailable falls back to db",
			shortCode: "abc123",
			repo: &mockLinkRepo{
				findByShortCodeLink: mustLink(
					10, 1, "https://example.com", "abc123",
				),
			},
			cache: &mockLinkCache{
				hitLink: newLinkWithState(
					10, 1, "https://cached.example.com", "abc123", false, nil,
				),
				hitFound: true,
			},
			queue:          &mockQueue{},
			wantErr:        "",
			wantShortCode:  "abc123",
			wantOrigURL:    "https://example.com",
			wantUserID:     1,
			wantFindCalled: true,
			wantSet:        true,
			wantPublish:    true,
		},
		{
			name:      "cache unavailable falls back to db",
			shortCode: "abc123",
			repo: &mockLinkRepo{
				findByShortCodeLink: mustLink(
					10, 1, "https://example.com", "abc123",
				),
			},
			cache: &mockLinkCache{
				hitErr: linkCache.ErrCacheUnavailable,
			},
			queue:          &mockQueue{},
			wantErr:        "",
			wantShortCode:  "abc123",
			wantOrigURL:    "https://example.com",
			wantUserID:     1,
			wantFindCalled: true,
			wantSet:        true,
			wantPublish:    true,
		},
		{
			name:      "set failure ignored",
			shortCode: "abc123",
			repo: &mockLinkRepo{
				findByShortCodeLink: mustLink(
					10, 1, "https://example.com", "abc123",
				),
			},
			cache: &mockLinkCache{
				setErr: errors.New("cache set failed"),
			},
			queue:          &mockQueue{},
			wantErr:        "",
			wantShortCode:  "abc123",
			wantOrigURL:    "https://example.com",
			wantUserID:     1,
			wantFindCalled: true,
			wantSet:        true,
			wantPublish:    true,
		},
		{
			name:      "publish failure ignored",
			shortCode: "abc123",
			repo: &mockLinkRepo{
				findByShortCodeLink: mustLink(
					10, 1, "https://example.com", "abc123",
				),
			},
			cache:          &mockLinkCache{},
			queue:          &mockQueue{publishErr: errors.New("publish failed")},
			wantErr:        "",
			wantShortCode:  "abc123",
			wantOrigURL:    "https://example.com",
			wantUserID:     1,
			wantFindCalled: true,
			wantSet:        true,
			wantPublish:    true,
		},
		{
			name:      "no queue configured",
			shortCode: "abc123",
			repo: &mockLinkRepo{
				findByShortCodeLink: mustLink(
					10, 1, "https://example.com", "abc123",
				),
			},
			cache:          &mockLinkCache{},
			queue:          nil,
			wantErr:        "",
			wantShortCode:  "abc123",
			wantOrigURL:    "https://example.com",
			wantUserID:     1,
			wantFindCalled: true,
			wantSet:        true,
			wantPublish:    false,
		},
		{
			name:           "link doesn't exist",
			shortCode:      "nope",
			repo:           &mockLinkRepo{},
			cache:          &mockLinkCache{},
			queue:          &mockQueue{},
			wantErr:        "link not found",
			wantFindCalled: true,
		},
		{
			name:      "inactive link",
			shortCode: "abc123",
			repo: &mockLinkRepo{
				findByShortCodeLink: newLinkWithState(
					10, 1, "https://example.com", "abc123", false, nil,
				),
			},
			cache:          &mockLinkCache{},
			queue:          &mockQueue{},
			wantErr:        "link is not available",
			wantFindCalled: true,
		},
		{
			name:      "expired link",
			shortCode: "abc123",
			repo: &mockLinkRepo{
				findByShortCodeLink: newLinkWithState(
					10, 1, "https://example.com", "abc123", true,
					pastTime(),
				),
			},
			cache:          &mockLinkCache{},
			queue:          &mockQueue{},
			wantErr:        "link is not available",
			wantFindCalled: true,
		},
		{
			name:      "repository error",
			shortCode: "abc123",
			repo: &mockLinkRepo{
				findByShortCodeErr: repoErr,
			},
			cache:          &mockLinkCache{},
			queue:          &mockQueue{},
			wantErr:        "repository failed",
			wantFindCalled: true,
		},
		{
			name:           "short code empty",
			shortCode:      "",
			repo:           &mockLinkRepo{},
			cache:          &mockLinkCache{},
			queue:          &mockQueue{},
			wantErr:        "short code cannot be empty",
			wantFindCalled: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cache linkCache.LinkCache
			if tt.cache != nil {
				cache = tt.cache
			}

			var queue messagequeue.Rabbitmq
			if tt.queue != nil {
				queue = tt.queue
			}

			uc := NewResolveShortLinkUseCase(tt.repo, cache, queue)

			link, err := uc.ResolveShortLink(tt.shortCode)

			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
				if link == nil {
					t.Fatal("expected a link, got nil")
				}
				if got := link.ShortCode(); got != tt.wantShortCode {
					t.Errorf("expected short code %q, got %q", tt.wantShortCode, got)
				}
				if got := link.OriginalURL(); got != tt.wantOrigURL {
					t.Errorf("expected original url %q, got %q", tt.wantOrigURL, got)
				}
				if got := link.UserID(); got != tt.wantUserID {
					t.Errorf("expected user id %d, got %d", tt.wantUserID, got)
				}
			} else {
				if err == nil {
					t.Fatalf("expected error %q, got nil", tt.wantErr)
				}
				if err.Error() != tt.wantErr {
					t.Fatalf("expected error %q, got %q", tt.wantErr, err.Error())
				}
			}

			if tt.repo.calledFindByShortCode != tt.wantFindCalled {
				t.Errorf("expected repo.FindByShortCode called = %v, got %v", tt.wantFindCalled, tt.repo.calledFindByShortCode)
			}

			if tt.cache != nil {
				if tt.cache.calledSet != tt.wantSet {
					t.Errorf("expected cache.Set called = %v, got %v", tt.wantSet, tt.cache.calledSet)
				}
				if tt.wantSet && tt.cache.setLink != nil {
					if tt.cache.setLink.ShortCode() != tt.shortCode {
						t.Errorf("expected cache.Set short code %q, got %q", tt.shortCode, tt.cache.setLink.ShortCode())
					}
					if tt.cache.setLink.OriginalURL() != tt.wantOrigURL {
						t.Errorf("expected cache.Set url %q, got %q", tt.wantOrigURL, tt.cache.setLink.OriginalURL())
					}
				}
			}

			if tt.queue != nil {
				if tt.queue.calledPublish != tt.wantPublish {
					t.Errorf("expected queue.Publish called = %v, got %v", tt.wantPublish, tt.queue.calledPublish)
				}
				if tt.wantPublish {
					assertClickPayload(t, tt.queue, tt.wantUserID, tt.shortCode, tt.wantOrigURL)
				}
			}
		})
	}
}

func assertClickPayload(t *testing.T, q *mockQueue, wantUserID int, wantCode, wantURL string) {
	t.Helper()

	if q.routingKey != messagequeue.RoutingKeyLinkClicked {
		t.Errorf("expected routing key %q, got %q", messagequeue.RoutingKeyLinkClicked, q.routingKey)
	}

	var event linkDTO.LinkClickedEvent
	if err := json.Unmarshal(q.payload, &event); err != nil {
		t.Fatalf("invalid event payload: %v", err)
	}

	if event.EventID == "" {
		t.Error("expected event_id to be set")
	}
	if event.UserID != wantUserID {
		t.Errorf("expected user_id %d, got %d", wantUserID, event.UserID)
	}
	if event.ShortCode != wantCode {
		t.Errorf("expected short_code %q, got %q", wantCode, event.ShortCode)
	}
	if event.OriginalURL != wantURL {
		t.Errorf("expected original_url %q, got %q", wantURL, event.OriginalURL)
	}
	if event.ClickedAt.IsZero() {
		t.Error("expected clicked_at to be set")
	}
}