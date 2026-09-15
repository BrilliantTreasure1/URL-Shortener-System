package report

import (
	"net/http"
	"strconv"

	"url-shortener/application/report"

	"github.com/gin-gonic/gin"
)

type ReportOverviewController struct {
	reportUsecase report.ReportOverviewUsecase
}

func NewReportOverviewController(reportUsecase report.ReportOverviewUsecase) *ReportOverviewController {
	return &ReportOverviewController{
		reportUsecase: reportUsecase,
	}
}

func (r *ReportOverviewController) Overview(c *gin.Context) {

	userIDRaw, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "user not authenticated",
		})
		return
	}

	userID, err := strconv.Atoi(strconv.FormatFloat(userIDRaw.(float64), 'f', -1, 64))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "invalid user id",
		})
		return
	}

	overview, err := r.reportUsecase.GetOverview(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, overview)
}