package messagequeue

import (
	"context"
	"errors"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

var ErrQueueUnavailable = errors.New("queue is unavailable")

const exchangeLinkEvents = "link.events"

type RabbitmqPublisher struct {
	connection *amqp.Connection
}

func NewRabbitmqPublisher(connection *amqp.Connection) *RabbitmqPublisher {
	return &RabbitmqPublisher{
		connection: connection,
	}
}

func (p *RabbitmqPublisher) Publish(routingKey string, payload []byte) error {
	if p.connection == nil {
		return ErrQueueUnavailable
	}

	channel, err := p.connection.Channel()
	if err != nil {
		return err
	}
	defer channel.Close()

	err = channel.ExchangeDeclare(
		exchangeLinkEvents,
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = channel.PublishWithContext(
		ctx,
		exchangeLinkEvents,
		routingKey,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Body:         payload,
		},
	)
	if err != nil {
		return err
	}

	return nil
}

var _ Rabbitmq = (*RabbitmqPublisher)(nil)