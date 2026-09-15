package report

import (
	"errors"
	"testing"

	reportRepo "url-shortener/repository/report"
)

func TestGetOverview(t *testing.T) {
	genericErr := errors.New("repository failed")

	tests := []struct {
		name            string
		userID          int
		repo            *mockReportRepo
		wantErr         string
		wantRepoCalled  bool
		wantTotalLinks  int64
		wantActiveLinks int64
		wantTotalClicks int64
	}{
		{
			name:   "returns overview counts",
			userID: 1,
			repo: &mockReportRepo{
				overview: &reportRepo.Overview{
					TotalLinks:  3,
					ActiveLinks: 2,
					TotalClicks: 10,
				},
			},
			wantErr:         "",
			wantRepoCalled:  true,
			wantTotalLinks:  3,
			wantActiveLinks: 2,
			wantTotalClicks: 10,
		},
		{
			name:           "repository error",
			userID:         1,
			repo:           &mockReportRepo{countsErr: genericErr},
			wantErr:        "repository failed",
			wantRepoCalled: true,
		},
		{
			name:   "invalid user id",
			userID: 0,
			repo:   &mockReportRepo{},
			wantErr: "invalid user id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := NewReportOverviewUsecase(tt.repo)

			reportDto, err := uc.GetOverview(tt.userID)

			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
				if reportDto == nil {
					t.Fatal("expected a dto, got nil")
				}
				if reportDto.TotalLinks != tt.wantTotalLinks {
					t.Errorf("expected total links %d, got %d", tt.wantTotalLinks, reportDto.TotalLinks)
				}
				if reportDto.ActiveLinks != tt.wantActiveLinks {
					t.Errorf("expected active links %d, got %d", tt.wantActiveLinks, reportDto.ActiveLinks)
				}
				if reportDto.TotalClicks != tt.wantTotalClicks {
					t.Errorf("expected total clicks %d, got %d", tt.wantTotalClicks, reportDto.TotalClicks)
				}
			} else {
				if err == nil {
					t.Fatalf("expected error %q, got nil", tt.wantErr)
				}
				if err.Error() != tt.wantErr {
					t.Fatalf("expected error %q, got %q", tt.wantErr, err.Error())
				}
			}

			if tt.repo.calledCounts != tt.wantRepoCalled {
				t.Errorf("expected repo.CountsByUserID called = %v, got %v", tt.wantRepoCalled, tt.repo.calledCounts)
			}
			if tt.wantRepoCalled && tt.repo.callUserID != tt.userID {
				t.Errorf("expected repo user id %d, got %d", tt.userID, tt.repo.callUserID)
			}
		})
	}
}