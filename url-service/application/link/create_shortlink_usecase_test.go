package link

import (
	"errors"
	"testing"
	"time"

	entities "url-shortener/entities/link"
)

type mockGenerator struct {
	code string
	err  error

	calledSeq int64
	calledLen int
	callCount int
}

func (m *mockGenerator) GenerateUnique(seq int64, randomLength int) (string, error) {
	m.calledSeq = seq
	m.calledLen = randomLength
	m.callCount++
	return m.code, m.err
}

func TestCreateShortLink(t *testing.T) {
	genErr := errors.New("generator failed")
	repoErr := errors.New("repository failed")

	tests := []struct {
		name           string
		userID         int
		url            string
		gen            *mockGenerator
		repo           *mockLinkRepo
		wantErr        string
		wantShortCode  string
		wantOrigURL    string
		wantSeq        int64
		wantLen        int
		wantNextCalled bool
		wantCreateCall bool
		wantGenCalls   int
	}{
		{
			name:   "successfully creates link",
			userID: 1,
			url:    "https://example.com",
			gen: &mockGenerator{
				code: "abc123",
			},
			repo: &mockLinkRepo{
				nextValue: 42,
				createLink: mustLink(
					10, 1, "https://example.com", "abc123",
				),
			},
			wantErr:        "",
			wantShortCode:  "abc123",
			wantOrigURL:    "https://example.com",
			wantSeq:        42,
			wantLen:        5,
			wantNextCalled: true,
			wantCreateCall: true,
			wantGenCalls:   1,
		},
		{
			name:   "generator error",
			userID: 1,
			url:    "https://example.com",
			gen: &mockGenerator{
				err: genErr,
			},
			repo: &mockLinkRepo{
				nextValue: 42,
			},
			wantErr:        "generator failed",
			wantNextCalled: true,
			wantCreateCall: false,
			wantGenCalls:   1,
		},
		{
			name:   "repository error",
			userID: 1,
			url:    "https://example.com",
			gen: &mockGenerator{
				code: "abc123",
			},
			repo: &mockLinkRepo{
				nextValue: 42,
				createErr: repoErr,
			},
			wantErr:        "repository failed",
			wantNextCalled: true,
			wantCreateCall: true,
			wantGenCalls:   1,
		},
		{
			name:    "invalid user id",
			userID:  0,
			url:     "https://example.com",
			gen:     &mockGenerator{},
			repo:    &mockLinkRepo{},
			wantErr: "invalid user id",
		},
		{
			name:    "empty original url",
			userID:  1,
			url:     "",
			gen:     &mockGenerator{},
			repo:    &mockLinkRepo{},
			wantErr: "original url cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := NewCreateShortLinkUseCase(tt.gen, tt.repo)

			link, err := uc.CreateShortLink(tt.userID, tt.url)

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
				if tt.gen.calledSeq != tt.wantSeq {
					t.Errorf("expected generator called with seq %d, got %d", tt.wantSeq, tt.gen.calledSeq)
				}
				if tt.gen.calledLen != tt.wantLen {
					t.Errorf("expected generator called with length %d, got %d", tt.wantLen, tt.gen.calledLen)
				}
			} else {
				if err == nil {
					t.Fatalf("expected error %q, got nil", tt.wantErr)
				}
				if err.Error() != tt.wantErr {
					t.Fatalf("expected error %q, got %q", tt.wantErr, err.Error())
				}
			}

			if tt.gen.callCount != tt.wantGenCalls {
				t.Errorf("expected %d generator calls, got %d", tt.wantGenCalls, tt.gen.callCount)
			}
			if tt.repo.calledNext != tt.wantNextCalled {
				t.Errorf("expected repo.NextCodeValue called = %v, got %v", tt.wantNextCalled, tt.repo.calledNext)
			}
			if tt.repo.calledCreate != tt.wantCreateCall {
				t.Errorf("expected repo.Create called = %v, got %v", tt.wantCreateCall, tt.repo.calledCreate)
			}
		})
	}
}

func mustLink(id int, userID int, originalURL string, shortCode string) *entities.Link {
	link, err := entities.NewLinkFromDatabase(
		&id,
		userID,
		originalURL,
		shortCode,
		time.Now(),
		nil,
		true,
	)
	if err != nil {
		panic(err)
	}
	return link
}