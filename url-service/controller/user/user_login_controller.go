package user

import (
	"url-shortener/application/user"
	"url-shortener/application/user/dto"
	"net/http"

	"github.com/gin-gonic/gin"
)

type UserLoginController struct{
	userUsecase user.LoginInterface
}

func NewUserLoginController(userUsecase user.LoginInterface) *UserLoginController{
	return &UserLoginController{
		userUsecase: userUsecase,
	}
}

func (u *UserLoginController) Login(c *gin.Context) {
	var request struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}

	token, err := u.userUsecase.Login(
		request.Email,
		request.Password,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.UserResponse{
		Token: token,
	})
}	
