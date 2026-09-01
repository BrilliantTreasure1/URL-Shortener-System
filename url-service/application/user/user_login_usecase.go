package user

import (
	"url-shortener/config"
	userRepo "url-shortener/repository/user"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type LoginUseCase struct {
	userRepo userRepo.UserRepository
}

func NewLoginUseCase(userRepo userRepo.UserRepository) *LoginUseCase {
	return &LoginUseCase{
		userRepo: userRepo,
	}
}

func (l *LoginUseCase) Login(
	email string,
	password string,
) (string, error) {

	user, err := l.userRepo.FindByEmail(email)
	if err != nil {
		return "", err
	}

	if user == nil {
		return "", errors.New("invalid email or password")
	}

	err = bcrypt.CompareHashAndPassword(
		[]byte(user.Password()),
		[]byte(password),
	)

	if err != nil {
		return "", errors.New("invalid email or password")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": *user.Id(),
		"email":   user.Email(),
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	})

	tokenString, err := token.SignedString([]byte(config.GetJWTSecret()))
	if err != nil {
		return "", errors.New("failed to generate token")
	}

	return tokenString, nil
}