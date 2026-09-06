package link

import (
	"errors"
	"testing"

	linkRepo "url-shortener/repository/link"
)

func TestDisableShortLink(t *testing.T) {
	genericErr := errors.New("repository failed")

	tests := []struct {
		name                string
		userID              int
		shortCode           string
		repo                *mockLinkRepo
		cache               *mockLinkCache
		wantErr             string
		wantAlreadyDisabled bool
		wantDisable         bool
		wantDisableUserID   int
		wantDisableCode     string
		wantDelete          bool
	}{
		{
			name:      "owner disables link",
			userID:    1,
			shortCode: "abc123",
			repo: &mockLinkRepo{
				disableLink: newLinkWithState(
					10, 1, "https://example.com", "abc123", false, nil,
				),
			},
			cache: &mockLinkCache{
				deleteErr: errors.New("cache delete failed"),
			},
			wantErr:           "",
			wantDisable:       true,
			wantDisableUserID: 1,
			wantDisableCode:   "abc123",
			wantDelete:        true,
		},
		{
			name:              "link doesn't exist",
			userID:            1,
			shortCode:         "nope",
			repo:              &mockLinkRepo{},
			cache:             &mockLinkCache{},
			wantErr:           "link not found",
			wantDisable:       true,
			wantDisableUserID: 1,
			wantDisableCode:   "nope",
		},
		{
			name:              "user doesn't own link",
			userID:            3,
			shortCode:         "abc123",
			repo:              &mockLinkRepo{},
			cache:             &mockLinkCache{},
			wantErr:           "link not found",
			wantDisable:       true,
			wantDisableUserID: 3,
			wantDisableCode:   "abc123",
		},
		{
			name:      "already disabled",
			userID:    1,
			shortCode: "abc123",
			repo: &mockLinkRepo{
				disableErr: linkRepo.ErrLinkAlreadyDisabled,
			},
			cache:              &mockLinkCache{},
			wantErr:             "link is already disabled",
			wantAlreadyDisabled: true,
			wantDisable:         true,
			wantDisableUserID:   1,
			wantDisableCode:     "abc123",
		},
		{
			name:      "repository error",
			userID:    1,
			shortCode: "abc123",
			repo: &mockLinkRepo{
				disableErr: genericErr,
			},
			cache:             &mockLinkCache{},
			wantErr:           "repository failed",
			wantDisable:       true,
			wantDisableUserID: 1,
			wantDisableCode:   "abc123",
		},
		{
			name:        "invalid user id",
			userID:      0,
			shortCode:   "abc123",
			repo:        &mockLinkRepo{},
			cache:       &mockLinkCache{},
			wantErr:     "invalid user id",
			wantDisable: false,
		},
		{
			name:        "short code empty",
			userID:      1,
			shortCode:   "",
			repo:        &mockLinkRepo{},
			cache:       &mockLinkCache{},
			wantErr:     "short code cannot be empty",
			wantDisable: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := NewDisableShortLinkUseCase(tt.repo, tt.cache)

			link, err := uc.DisableShortLink(tt.userID, tt.shortCode)

			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
				if link == nil {
					t.Fatal("expected a link, got nil")
				}
				if link.IsActive() {
					t.Error("expected disabled link, got is_active = true")
				}
			} else {
				if err == nil {
					t.Fatalf("expected error %q, got nil", tt.wantErr)
				}
				if tt.wantAlreadyDisabled {
					if !errors.Is(err, ErrLinkAlreadyDisabled) {
						t.Errorf("expected ErrLinkAlreadyDisabled, got %v", err)
					}
				}
				if err.Error() != tt.wantErr {
					t.Fatalf("expected error %q, got %q", tt.wantErr, err.Error())
				}
			}

			if tt.repo.calledDisable != tt.wantDisable {
				t.Errorf("expected repo.DisableLink called = %v, got %v", tt.wantDisable, tt.repo.calledDisable)
			}
			if tt.wantDisableUserID != 0 && tt.repo.callDisableUserID != tt.wantDisableUserID {
				t.Errorf("expected repo.DisableLink userID %d, got %d", tt.wantDisableUserID, tt.repo.callDisableUserID)
			}
			if tt.wantDisableCode != "" && tt.repo.callDisableShortCode != tt.wantDisableCode {
				t.Errorf("expected repo.DisableLink shortCode %q, got %q", tt.wantDisableCode, tt.repo.callDisableShortCode)
			}

			if tt.cache != nil {
				if tt.cache.calledDelete != tt.wantDelete {
					t.Errorf("expected cache.Delete called = %v, got %v", tt.wantDelete, tt.cache.calledDelete)
				}
				if tt.wantDelete && tt.cache.deleteCode != tt.shortCode {
					t.Errorf("expected cache.Delete short code %q, got %q", tt.shortCode, tt.cache.deleteCode)
				}
			}
		})
	}
}
