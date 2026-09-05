package link

import (
	"errors"
	"testing"

	entities "url-shortener/entities/link"
)

func TestListShortLink(t *testing.T) {
	repoErr := errors.New("repository failed")

	tests := []struct {
		name           string
		userID         int
		offset         int
		limit          int
		repo           *mockLinkRepo
		wantErr        string
		wantLen        int
		wantTotal      int64
		wantFirstCode  string
		wantList       bool
		wantListUserID int
		wantListOffset int
		wantListLimit  int
	}{
		{
			name:   "returns user's links",
			userID: 1,
			offset: 0,
			limit:  10,
			repo: &mockLinkRepo{
				listLinks: []*entities.Link{
					newLinkWithState(1, 1, "https://one.com", "code1", true, nil),
					newLinkWithState(2, 1, "https://two.com", "code2", true, nil),
				},
				listTotal: 2,
			},
			wantErr:        "",
			wantLen:        2,
			wantTotal:      2,
			wantFirstCode:  "code1",
			wantList:       true,
			wantListUserID: 1,
			wantListOffset: 0,
			wantListLimit:  10,
		},
		{
			name:   "pagination",
			userID: 1,
			offset: 5,
			limit:  3,
			repo: &mockLinkRepo{
				listLinks: []*entities.Link{
					newLinkWithState(6, 1, "https://six.com", "code6", true, nil),
					newLinkWithState(7, 1, "https://seven.com", "code7", true, nil),
					newLinkWithState(8, 1, "https://eight.com", "code8", true, nil),
				},
				listTotal: 8,
			},
			wantErr:        "",
			wantLen:        3,
			wantTotal:      8,
			wantFirstCode:  "code6",
			wantList:       true,
			wantListUserID: 1,
			wantListOffset: 5,
			wantListLimit:  3,
		},
		{
			name:   "empty result",
			userID: 1,
			offset: 0,
			limit:  10,
			repo: &mockLinkRepo{
				listLinks: []*entities.Link{},
				listTotal: 0,
			},
			wantErr:        "",
			wantLen:        0,
			wantTotal:      0,
			wantList:       true,
			wantListUserID: 1,
			wantListOffset: 0,
			wantListLimit:  10,
		},
		{
			name:   "repository error",
			userID: 1,
			offset: 0,
			limit:  10,
			repo: &mockLinkRepo{
				listErr: repoErr,
			},
			wantErr:        "repository failed",
			wantList:       true,
			wantListUserID: 1,
			wantListOffset: 0,
			wantListLimit:  10,
		},
		{
			name:    "invalid user id",
			userID:  0,
			offset:  0,
			limit:   10,
			repo:    &mockLinkRepo{},
			wantErr: "invalid user id",
		},
		{
			name:    "negative offset",
			userID:  1,
			offset:  -1,
			limit:   10,
			repo:    &mockLinkRepo{},
			wantErr: "offset cannot be negative",
		},
		{
			name:    "invalid limit",
			userID:  1,
			offset:  0,
			limit:   0,
			repo:    &mockLinkRepo{},
			wantErr: "limit must be greater than zero",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := NewListShortLinkUseCase(tt.repo)

			links, total, err := uc.ListShortLink(tt.userID, tt.offset, tt.limit)

			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
				if len(links) != tt.wantLen {
					t.Errorf("expected %d links, got %d", tt.wantLen, len(links))
				}
				if total != tt.wantTotal {
					t.Errorf("expected total %d, got %d", tt.wantTotal, total)
				}
				if tt.wantFirstCode != "" && len(links) > 0 {
					if got := links[0].ShortCode(); got != tt.wantFirstCode {
						t.Errorf("expected first link short code %q, got %q", tt.wantFirstCode, got)
					}
				}
			} else {
				if err == nil {
					t.Fatalf("expected error %q, got nil", tt.wantErr)
				}
				if err.Error() != tt.wantErr {
					t.Fatalf("expected error %q, got %q", tt.wantErr, err.Error())
				}
			}

			if tt.repo.calledList != tt.wantList {
				t.Errorf("expected repo.ListByUserID called = %v, got %v", tt.wantList, tt.repo.calledList)
			}
			if tt.wantListUserID != 0 && tt.repo.callListUserID != tt.wantListUserID {
				t.Errorf("expected repo.ListByUserID userID %d, got %d", tt.wantListUserID, tt.repo.callListUserID)
			}
			if tt.repo.callListOffset != tt.wantListOffset {
				t.Errorf("expected repo.ListByUserID offset %d, got %d", tt.wantListOffset, tt.repo.callListOffset)
			}
			if tt.repo.callListLimit != tt.wantListLimit {
				t.Errorf("expected repo.ListByUserID limit %d, got %d", tt.wantListLimit, tt.repo.callListLimit)
			}
		})
	}
}
