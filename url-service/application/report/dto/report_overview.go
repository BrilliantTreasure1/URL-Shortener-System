package dto

type ReportOverviewDto struct {
	TotalLinks  int64 `json:"total_links"`
	ActiveLinks int64 `json:"active_links"`
	TotalClicks int64 `json:"total_clicks"`
}