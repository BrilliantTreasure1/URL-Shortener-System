package user

import (
	"url-shortener/application/user"
	"net/http"

	"github.com/gin-gonic/gin"
)

type UserController struct{
	userUsecase user.RegisterInterface
}

func NewUserController(userUsecase user.RegisterInterface) *UserController{
	return &UserController{
		userUsecase: userUsecase,
	}
}

func (u *UserController) Register(c *gin.Context) {
	var request struct {
		Username    string `json:"username"`
		Email       string `json:"email"`
		PhoneNumber string `json:"phonenumber"`
		Password    string `json:"password"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}

	user, err := u.userUsecase.Register(
		request.Username,
		request.Email,
		request.PhoneNumber,
		request.Password,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, user.ToResponse())
}

