package user

import (
	"url-shortener/entities/user"
)

type RegisterInterface interface {
	Register(username string , email string , phonenumber string , password string) (*entities.User , error)
}

type LoginInterface interface {
	Login(email string , password string) (string , error)
}