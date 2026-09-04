package link

import (
	"net/http"
	"strconv"

	"url-shortener/application/link"

	"github.com/gin-gonic/gin"
)

type CreateLinkController struct {
	linkUsecase link.CreateShortLinkUseCase
}

func NewCreateLinkController(linkUsecase link.CreateShortLinkUseCase) *CreateLinkController {
	return &CreateLinkController{
		linkUsecase: linkUsecase,
	}
}

func (cl *CreateLinkController) Create(c *gin.Context) {
	var request struct {
		OriginalURL string `json:"original_url"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}

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

	link, err := cl.linkUsecase.CreateShortLink(userID, request.OriginalURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, link.ToResponse())
}


