package container

import (

	"database/sql"
	"time"

	applicationLink "url-shortener/application/link"
	controllerLink "url-shortener/controller/link"
	repositoryLink "url-shortener/repository/link"
	repositoryCache "url-shortener/repository/cache"
	messagequeue "url-shortener/message-queue"
	"url-shortener/service/shortcode-generator"

	"github.com/redis/go-redis/v9"
	amqp "github.com/rabbitmq/amqp091-go"
)

type LinkContainer struct {
	CreateLinkController  *controllerLink.CreateLinkController
	ResolveLinkController *controllerLink.ResolveLinkController
	ListLinkController    *controllerLink.ListLinkController
	DisableLinkController *controllerLink.DisableLinkController
}

func NewLinkContainer(db *sql.DB, redisClient *redis.Client, rabbitMQConnection *amqp.Connection, cacheTTL time.Duration) (*LinkContainer, error) {

	linkRepository := repositoryLink.NewLinkRepositoryPostgresql(db)
	shortcodeGenerator := shortcodegenerator.NewUniqueGenerator()

	linkCache := repositoryCache.NewRedisLinkCache(redisClient, cacheTTL)
	queue := messagequeue.NewRabbitmqPublisher(rabbitMQConnection)

	createShortLinkUseCase := applicationLink.NewCreateShortLinkUseCase(
		shortcodeGenerator,
		linkRepository,
	)

	createLinkController := controllerLink.NewCreateLinkController(
		*createShortLinkUseCase,
	)

	resolveShortLinkUseCase := applicationLink.NewResolveShortLinkUseCase(
		linkRepository,
		linkCache,
		queue,
	)

	resolveLinkController := controllerLink.NewResolveLinkController(
		*resolveShortLinkUseCase,
	)

	listShortLinkUseCase := applicationLink.NewListShortLinkUseCase(
		linkRepository,
	)

	listLinkController := controllerLink.NewListLinkController(
		*listShortLinkUseCase,
	)

	disableShortLinkUseCase := applicationLink.NewDisableShortLinkUseCase(
		linkRepository,
		linkCache,
	)

	disableLinkController := controllerLink.NewDisableLinkController(
		*disableShortLinkUseCase,
	)

	return &LinkContainer{
		CreateLinkController:  createLinkController,
		ResolveLinkController: resolveLinkController,
		ListLinkController:    listLinkController,
		DisableLinkController: disableLinkController,
	}, nil
}