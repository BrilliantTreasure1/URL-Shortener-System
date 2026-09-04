package main

import (
	"url-shortener/container"
	"url-shortener/middleware"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {

	app, err := container.NewContainer()
	if err != nil {
		log.Fatal("failed to initialize application:", err)
	}

	router := gin.Default()

	router.POST(
		"/users/register",
		app.User.UserController.Register,
	)

	router.POST(
		"/users/login",
		app.User.UserLoginController.Login,
	)

	protected := router.Group("/")
	protected.Use(middleware.AuthMiddleware())
	{
		protected.GET("/ping", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "pong",
				"user_id": c.GetString("user_id"),
			})
		})

		protected.POST("/links", app.Link.CreateLinkController.Create)
	}

	log.Println("Server running on :8080")

	if err := router.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}