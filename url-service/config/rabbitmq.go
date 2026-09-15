package config

import (
	"fmt"
	"net"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	rabbitMQDialTimeout = 5 * time.Second
	rabbitMQMaxRetries  = 15
	rabbitMQRetryDelay  = 3 * time.Second
)

func dialRabbitMQ(amqpURL string) (*amqp.Connection, error) {

	var lastErr error

	for attempt := 1; attempt <= rabbitMQMaxRetries; attempt++ {
		connection, err := amqp.DialConfig(amqpURL, amqp.Config{
			Dial: (&net.Dialer{
				Timeout: rabbitMQDialTimeout,
			}).Dial,
		})
		if err == nil {
			return connection, nil
		}

		lastErr = err

		if attempt < rabbitMQMaxRetries {
			time.Sleep(rabbitMQRetryDelay)
		}
	}

	return nil, lastErr
}

func NewRabbitMQConnection() (*amqp.Connection, error) {

	amqpURL := fmt.Sprintf(
		"amqp://%s:%s@%s:%s/%s",
		getEnv("RABBITMQUSER", "guest"),
		getEnv("RABBITMQPASSWORD", "guest"),
		getEnv("RABBITMQHOST", "localhost"),
		getEnv("RABBITMQPORT", "5672"),
		getEnv("RABBITMQVHOST", ""),
	)

	return dialRabbitMQ(amqpURL)
}