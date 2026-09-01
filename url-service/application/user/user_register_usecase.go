package user

import (
	"url-shortener/entities/user"
	"golang.org/x/crypto/bcrypt"
	"errors"
	"url-shortener/repository/user"

)


type RegisterUseCase struct{
	userRepo user.UserRepository
}

func NewRegisterUseCase(userRepo user.UserRepository) *RegisterUseCase{
	return &RegisterUseCase{
        userRepo: userRepo,
    }
}

func (r *RegisterUseCase) Register(  
	username string,
    email string,
    phoneNumber string,
    password string,
)(*entities.User , error){
	existing , err := r.userRepo.FindByEmail(email)

	if err != nil {
    return nil, err
}

	if existing != nil{
		return nil , errors.New("user already registered")
	}

	hash , err := bcrypt.GenerateFromPassword([]byte(password),bcrypt.DefaultCost)

	if err != nil {
        return nil, err
    }

	user , err := entities.NewUser(nil, username , email , phoneNumber , string(hash))

	   if err != nil {
        return nil, err
    }

	result , err := r.userRepo.Register(user)

	 if err != nil {
        return nil, err
    }

	return result , err
}