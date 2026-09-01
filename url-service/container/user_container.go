package container

import (

	"database/sql"

	applicationUser "url-shortener/application/user"
	controllerUser "url-shortener/controller/user"
	repositoryUser "url-shortener/repository/user"
)

type UserContainer struct {
	UserController    *controllerUser.UserController
	UserLoginController *controllerUser.UserLoginController
}

func NewUserContainer(db *sql.DB) (*UserContainer, error) {

	userRepository := repositoryUser.NewUserRepositoryPostgresql(db)

	registerUseCase := applicationUser.NewRegisterUseCase(
		userRepository,
	)

	loginUseCase := applicationUser.NewLoginUseCase(
		userRepository,
	)

	userController := controllerUser.NewUserController(
		registerUseCase,
	)

	userLoginController := controllerUser.NewUserLoginController(
		loginUseCase,
	)

	return &UserContainer{
		UserController:    userController,
		UserLoginController: userLoginController,
	}, nil
}