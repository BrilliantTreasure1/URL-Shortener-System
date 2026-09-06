package link

import (
	"errors"
	"testing"

	linkCache "url-shortener/repository/cache"
)

func TestResolveShortLink(t *testing.T) {
	repoErr := errors.New("repository failed")

	tests := []struct {
		name           string
		shortCode      string
		repo           *mockLinkRepo
		cache          *mockLinkCache
		wantErr        string
		wantShortCode  string
		wantOrigURL    string
		wantFindCalled bool
		wantSet        bool
		wantSetURL     string
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
			wantErr:        "",
			wantShortCode:  "abc123",
			wantOrigURL:    "https://example.com",
			wantFindCalled: true,
			wantSet:        true,
			wantSetURL:     "https://example.com",
		},
		{
			name:      "serves from cache",
			shortCode: "abc123",
			repo:      &mockLinkRepo{},
			cache: &mockLinkCache{
				hitURL:   "https://cached.example.com",
				hitFound: true,
			},
			wantErr:        "",
			wantShortCode:  "abc123",
			wantOrigURL:    "https://cached.example.com",
			wantFindCalled: false,
			wantSet:        false,
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
			wantErr:        "",
			wantShortCode:  "abc123",
			wantOrigURL:    "https://example.com",
			wantFindCalled: true,
			wantSet:        true,
			wantSetURL:     "https://example.com",
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
			wantErr:        "",
			wantShortCode:  "abc123",
			wantOrigURL:    "https://example.com",
			wantFindCalled: true,
			wantSet:        true,
			wantSetURL:     "https://example.com",
		},
		{
			name:      "no cache configured",
			shortCode: "abc123",
			repo: &mockLinkRepo{
				findByShortCodeLink: mustLink(
					10, 1, "https://example.com", "abc123",
				),
			},
			cache:          nil,
			wantErr:        "",
			wantShortCode:  "abc123",
			wantOrigURL:    "https://example.com",
			wantFindCalled: true,
		},
		{
			name:           "link doesn't exist",
			shortCode:      "nope",
			repo:           &mockLinkRepo{},
			cache:          &mockLinkCache{},
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
			wantErr:        "repository failed",
			wantFindCalled: true,
		},
		{
			name:           "short code empty",
			shortCode:      "",
			repo:           &mockLinkRepo{},
			cache:          &mockLinkCache{},
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

			uc := NewResolveShortLinkUseCase(tt.repo, cache)

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
				if tt.wantSetURL != "" && tt.cache.setURL != tt.wantSetURL {
					t.Errorf("expected cache.Set url %q, got %q", tt.wantSetURL, tt.cache.setURL)
				}
				if tt.wantSet && tt.cache.setCode != tt.shortCode {
					t.Errorf("expected cache.Set short code %q, got %q", tt.shortCode, tt.cache.setCode)
				}
			}
		})
	}
}