package report

import (
	"errors"

	reportDTO "url-shortener/application/report/dto"
	reportRepo "url-shortener/repository/report"
)

type ReportOverviewUsecase struct {
	reportRepo reportRepo.ReportRepository
}

func NewReportOverviewUsecase(
	reportRepo reportRepo.ReportRepository,
) *ReportOverviewUsecase {
	return &ReportOverviewUsecase{
		reportRepo: reportRepo,
	}
}

func (r *ReportOverviewUsecase) GetOverview(userID int) (*reportDTO.ReportOverviewDto, error) {

	if userID <= 0 {
		return nil, errors.New("invalid user id")
	}

	overview, err := r.reportRepo.CountsByUserID(userID)
	if err != nil {
		return nil, err
	}

	return &reportDTO.ReportOverviewDto{
		TotalLinks:  overview.TotalLinks,
		ActiveLinks: overview.ActiveLinks,
		TotalClicks: overview.TotalClicks,
	}, nil
}