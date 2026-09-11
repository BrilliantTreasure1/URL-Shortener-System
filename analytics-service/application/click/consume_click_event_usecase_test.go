package click

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	clickDTO "analytics-service/application/click/dto"
)

func TestConsumeClickEvent(t *testing.T) {
	genericErr := errors.New("repository failed")

	validPayload := func() []byte {
		payload, _ := json.Marshal(clickDTO.LinkClickedEvent{
			EventID:     "event-1",
			UserID:      1,
			ShortCode:   "abc123",
			OriginalURL: "https://example.com",
			ClickedAt:   time.Now(),
		})
		return payload
	}()

	invalidEntityPayload := []byte(`{"event_id":"event-2","user_id":0,"short_code":"abc123","original_url":"https://example.com","clicked_at":"2026-09-11T10:00:00Z"}`)

	tests := []struct {
		name            string
		payload         []byte
		repo            *mockClickRepo
		wantErr         string
		wantErrPayload  bool
		wantRepoCalled  bool
		wantEventID     string
		wantUserID      int
		wantShortCode   string
		wantOriginalURL string
	}{
		{
			name:            "valid payload saved",
			payload:         validPayload,
			repo:            &mockClickRepo{saveResult: true},
			wantErr:         "",
			wantRepoCalled:  true,
			wantEventID:     "event-1",
			wantUserID:      1,
			wantShortCode:   "abc123",
			wantOriginalURL: "https://example.com",
		},
		{
			name:           "invalid json",
			payload:        []byte("not-json"),
			repo:           &mockClickRepo{},
			wantErr:        "invalid click event payload",
			wantErrPayload: true,
		},
		{
			name:           "invalid entity data",
			payload:        invalidEntityPayload,
			repo:           &mockClickRepo{},
			wantErr:        "invalid click event payload",
			wantErrPayload: true,
		},
		{
			name:           "duplicate event silently ignored",
			payload:        validPayload,
			repo:           &mockClickRepo{saveResult: false},
			wantErr:        "",
			wantRepoCalled: true,
			wantEventID:    "event-1",
		},
		{
			name:           "repository error",
			payload:        validPayload,
			repo:           &mockClickRepo{saveErr: genericErr},
			wantErr:        "repository failed",
			wantRepoCalled: true,
			wantEventID:    "event-1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := NewConsumeClickEventUseCase(tt.repo)

			err := uc.Handle(tt.payload)

			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
			} else {
				if err == nil {
					t.Fatalf("expected error %q, got nil", tt.wantErr)
				}
				if tt.wantErrPayload {
					if !errors.Is(err, ErrInvalidPayload) {
						t.Errorf("expected ErrInvalidPayload, got %v", err)
					}
				}
				if err.Error() != tt.wantErr {
					t.Fatalf("expected error %q, got %q", tt.wantErr, err.Error())
				}
			}

			if tt.repo.calledSave != tt.wantRepoCalled {
				t.Errorf("expected repo.Save called = %v, got %v", tt.wantRepoCalled, tt.repo.calledSave)
			}

			if tt.wantRepoCalled {
				if tt.repo.savedEvent == nil {
					t.Fatal("expected a saved event, got nil")
				}
				if tt.repo.savedEvent.EventID() != tt.wantEventID {
					t.Errorf("expected event id %q, got %q", tt.wantEventID, tt.repo.savedEvent.EventID())
				}
				if tt.wantUserID != 0 && tt.repo.savedEvent.UserID() != tt.wantUserID {
					t.Errorf("expected user id %d, got %d", tt.wantUserID, tt.repo.savedEvent.UserID())
				}
				if tt.wantShortCode != "" && tt.repo.savedEvent.ShortCode() != tt.wantShortCode {
					t.Errorf("expected short code %q, got %q", tt.wantShortCode, tt.repo.savedEvent.ShortCode())
				}
				if tt.wantOriginalURL != "" && tt.repo.savedEvent.OriginalURL() != tt.wantOriginalURL {
					t.Errorf("expected original url %q, got %q", tt.wantOriginalURL, tt.repo.savedEvent.OriginalURL())
				}
			}
		})
	}
}