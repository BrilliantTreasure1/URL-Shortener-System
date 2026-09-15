package container

import (
	"database/sql"

	applicationReport "url-shortener/application/report"
	controllerReport "url-shortener/controller/report"
	repositoryReport "url-shortener/repository/report"
)

type ReportContainer struct {
	ReportOverviewController *controllerReport.ReportOverviewController
}

func NewReportContainer(db *sql.DB) (*ReportContainer, error) {

	reportRepository := repositoryReport.NewReportRepositoryPostgresql(db)

	reportOverviewUsecase := applicationReport.NewReportOverviewUsecase(
		reportRepository,
	)

	reportOverviewController := controllerReport.NewReportOverviewController(
		*reportOverviewUsecase,
	)

	return &ReportContainer{
		ReportOverviewController: reportOverviewController,
	}, nil
}