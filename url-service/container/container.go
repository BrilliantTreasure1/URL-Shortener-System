package container

import(
	"url-shortener/config"
)

type Container struct {
	User *UserContainer
}

func NewContainer() (*Container, error) {

		db, err := config.NewDatabase()
	if err != nil {
		return nil, err
	}

	userContainer, err := NewUserContainer(db)
	if err != nil {
		return nil, err
	}

	return &Container{
		User: userContainer,
	}, nil
}