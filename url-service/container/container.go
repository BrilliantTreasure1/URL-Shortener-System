package container

import(
	"url-shortener/config"
)

type Container struct {
	User *UserContainer
	Link *LinkContainer
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

	linkContainer, err := NewLinkContainer(db)
	if err != nil {
		return nil, err
	}

	return &Container{
		User: userContainer,
		Link: linkContainer,
	}, nil
}