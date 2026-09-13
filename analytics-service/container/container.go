package container

import (
	"database/sql"
	"log"

	"analytics-service/config"
	"analytics-service/database"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Container struct {
	DB    *sql.DB
	MQ    *amqp.Connection
	Click *ClickContainer
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

	rabbitMQConnection, err := config.NewRabbitMQConnection()
	if err != nil {
		log.Printf("warn: rabbitmq unavailable, consumer disabled: %v", err)
		rabbitMQConnection = nil
	}

	clickContainer := NewClickContainer(db, rabbitMQConnection)

	return &Container{
		DB:    db,
		MQ:    rabbitMQConnection,
		Click: clickContainer,
	}, nil
}

func (c *Container) Close() {
	if c.MQ != nil {
		c.MQ.Close()
	}

	if c.DB != nil {
		c.DB.Close()
	}
}