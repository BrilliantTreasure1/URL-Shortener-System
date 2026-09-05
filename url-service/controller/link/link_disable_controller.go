package link

import (
	"errors"
	"net/http"
	"strconv"

	"url-shortener/application/link"

	"github.com/gin-gonic/gin"
)

type DisableLinkController struct {
	linkUsecase link.DisableShortLinkUseCase
}

func NewDisableLinkController(linkUsecase link.DisableShortLinkUseCase) *DisableLinkController {
	return &DisableLinkController{
		linkUsecase: linkUsecase,
	}
}

func (dl *DisableLinkController) Disable(c *gin.Context) {

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

	shortCode := c.Param("short_code")

	linkEntity, err := dl.linkUsecase.DisableShortLink(userID, shortCode)
	if err != nil {
		switch {
		case errors.Is(err, link.ErrLinkAlreadyDisabled):
			c.JSON(http.StatusConflict, gin.H{
				"error": err.Error(),
			})
		case err.Error() == "link not found":
			c.JSON(http.StatusNotFound, gin.H{
				"error": err.Error(),
			})
		case err.Error() == "short code cannot be empty":
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
		}
		return
	}

	c.JSON(http.StatusOK, linkEntity.ToResponse())
}