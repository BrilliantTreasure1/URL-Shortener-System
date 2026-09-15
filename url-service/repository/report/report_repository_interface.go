package report

type ReportRepository interface {
	CountsByUserID(userID int) (*Overview, error)
}