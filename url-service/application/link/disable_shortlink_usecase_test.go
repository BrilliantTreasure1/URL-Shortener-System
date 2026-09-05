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
		wantErr             string
		wantAlreadyDisabled bool
		wantDisable         bool
		wantDisableUserID   int
		wantDisableCode     string
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
			wantErr:           "",
			wantDisable:       true,
			wantDisableUserID: 1,
			wantDisableCode:   "abc123",
		},
		{
			name:              "link doesn't exist",
			userID:            1,
			shortCode:         "nope",
			repo:              &mockLinkRepo{},
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
			wantErr:     "invalid user id",
			wantDisable: false,
		},
		{
			name:        "short code empty",
			userID:      1,
			shortCode:   "",
			repo:        &mockLinkRepo{},
			wantErr:     "short code cannot be empty",
			wantDisable: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := NewDisableShortLinkUseCase(tt.repo)

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
		})
	}
}
