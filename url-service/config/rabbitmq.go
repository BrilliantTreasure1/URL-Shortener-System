package config

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

func NewRabbitMQConnection() (*amqp.Connection, error) {

	amqpURL := fmt.Sprintf(
		"amqp://%s:%s@%s:%s/%s",
		getEnv("RABBITMQUSER", "guest"),
		getEnv("RABBITMQPASSWORD", "guest"),
		getEnv("RABBITMQHOST", "localhost"),
		getEnv("RABBITMQPORT", "5672"),
		getEnv("RABBITMQVHOST", ""),
	)

	connection, err := amqp.Dial(amqpURL)
	if err != nil {
		return nil, err
	}

	return connection, nil
}