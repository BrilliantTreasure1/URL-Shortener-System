package report

import (
	reportRepo "url-shortener/repository/report"
)

type mockReportRepo struct {
	overview     *reportRepo.Overview
	countsErr    error
	calledCounts bool
	callUserID   int
}

func (m *mockReportRepo) CountsByUserID(userID int) (*reportRepo.Overview, error) {
	m.calledCounts = true
	m.callUserID = userID
	return m.overview, m.countsErr
}

var _ reportRepo.ReportRepository = (*mockReportRepo)(nil)