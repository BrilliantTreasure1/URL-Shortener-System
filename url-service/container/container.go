package container

import (
	"url-shortener/config"
	"url-shortener/database"
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

	if err := database.Migrate(db); err != nil {
		db.Close()
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