package link

import (
	"errors"
	"testing"
)

func TestResolveShortLink(t *testing.T) {
	repoErr := errors.New("repository failed")

	tests := []struct {
		name           string
		shortCode      string
		repo           *mockLinkRepo
		wantErr        string
		wantShortCode  string
		wantOrigURL    string
		wantFindCalled bool
	}{
		{
			name:      "finds link",
			shortCode: "abc123",
			repo: &mockLinkRepo{
				findByShortCodeLink: mustLink(
					10, 1, "https://example.com", "abc123",
				),
			},
			wantErr:        "",
			wantShortCode:  "abc123",
			wantOrigURL:    "https://example.com",
			wantFindCalled: true,
		},
		{
			name:           "link doesn't exist",
			shortCode:      "nope",
			repo:           &mockLinkRepo{},
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
			wantErr:        "link is not available",
			wantFindCalled: true,
		},
		{
			name:      "repository error",
			shortCode: "abc123",
			repo: &mockLinkRepo{
				findByShortCodeErr: repoErr,
			},
			wantErr:        "repository failed",
			wantFindCalled: true,
		},
		{
			name:           "short code empty",
			shortCode:      "",
			repo:           &mockLinkRepo{},
			wantErr:        "short code cannot be empty",
			wantFindCalled: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := NewResolveShortLinkUseCase(tt.repo)

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
		})
	}
}
