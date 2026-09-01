package user

import "url-shortener/entities/user"

type UserRepository interface {
    FindByEmail(email string) (*entities.User, error)
    Register(user *entities.User) (*entities.User, error)
}