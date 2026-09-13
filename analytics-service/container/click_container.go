package container

import (
	"database/sql"

	applicationClick "analytics-service/application/click"
	messagequeue "analytics-service/message-queue"
	repositoryClick "analytics-service/repository/click-event"

	amqp "github.com/rabbitmq/amqp091-go"
)

type ClickContainer struct {
	UseCase  *applicationClick.ConsumeClickEventUseCase
	Consumer messagequeue.Consumer
}

func NewClickContainer(db *sql.DB, rabbitMQConnection *amqp.Connection) *ClickContainer {

	clickRepository := repositoryClick.NewClickEventRepositoryPostgresql(db)

	consumeClickEventUseCase := applicationClick.NewConsumeClickEventUseCase(
		clickRepository,
	)

	consumer := messagequeue.NewRabbitmqConsumer(
		rabbitMQConnection,
	)

	return &ClickContainer{
		UseCase:  consumeClickEventUseCase,
		Consumer: consumer,
	}
}