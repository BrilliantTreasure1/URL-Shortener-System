package link

import (
	"net/http"
	"strconv"

	"url-shortener/application/link"
	"url-shortener/application/link/dto"

	"github.com/gin-gonic/gin"
)

type ListLinkController struct {
	linkUsecase link.ListShortLinkUseCase
}

func NewListLinkController(linkUsecase link.ListShortLinkUseCase) *ListLinkController {
	return &ListLinkController{
		linkUsecase: linkUsecase,
	}
}

func (ll *ListLinkController) List(c *gin.Context) {

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

	offset, err := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if err != nil || offset < 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "offset must be a non-negative integer",
		})
		return
	}

	limit, err := strconv.Atoi(c.DefaultQuery("limit", "15"))
	if err != nil || limit <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "limit must be a positive integer",
		})
		return
	}

	links, total, err := ll.linkUsecase.ListShortLink(userID, offset, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	responses := make([]dto.LinkResponse, 0, len(links))
	for _, link := range links {
		responses = append(responses, link.ToResponse())
	}

	c.JSON(http.StatusOK, gin.H{
		"links": responses,
		"total": total,
	})
}