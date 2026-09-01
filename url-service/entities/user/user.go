package entities

import (
    "errors"
    "strings"
	"url-shortener/application/user/dto"
)

type User struct {
    id          *int
    username    string
    email       string
    phonenumber string
    password    string
}

func NewUser(
    id *int,
    username string,
    email string,
    phonenumber string,
    password string,
) (*User, error) {

    if len(username) < 5 {
        return nil, errors.New("username is too short")
    }

    if !strings.Contains(email, "@") {
        return nil, errors.New("invalid email")
    }

    if len(phonenumber) < 11 {
        return nil, errors.New("phone number is too short")
    }

    if len(password) < 8 {
        return nil, errors.New("password too short")
    }

    return &User{
        id:          id,
        username:    username,
        email:       email,
        phonenumber: phonenumber,
        password:    password,
    }, nil
}

func (u *User) Id() *int{
	return u.id
}

func (u *User) Username() string{
	return u.username
}

func (u *User) Email() string {
	return  u.email
}
func (u *User) Phonenumber() string{
	return u.phonenumber
}

func (u *User) Password() string{
	return u.password
}

func (u *User) ToResponse() dto.UserResponse {
	return dto.UserResponse{
		ID:          *u.id,
		Username:    u.username,
		Email:       u.email,
		PhoneNumber: u.phonenumber,
	}
}