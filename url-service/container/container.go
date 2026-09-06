package container

import (
	"log"

	"url-shortener/config"
	"url-shortener/database"

	"github.com/redis/go-redis/v9"
)

type Container struct {
	User  *UserContainer
	Link  *LinkContainer
	Redis *redis.Client
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

	redisClient, err := config.NewRedisClient()
	if err != nil {
		log.Printf("warn: redis unavailable, continuing without cache: %v", err)
		redisClient = nil
	}

	userContainer, err := NewUserContainer(db)
	if err != nil {
		return nil, err
	}

	linkContainer, err := NewLinkContainer(db, redisClient, config.NewCacheTTL())
	if err != nil {
		return nil, err
	}

	return &Container{
		User:  userContainer,
		Link:  linkContainer,
		Redis: redisClient,
	}, nil
}