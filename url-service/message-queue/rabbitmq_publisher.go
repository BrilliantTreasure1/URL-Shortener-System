package messagequeue

import (
	"context"
	"errors"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

var ErrQueueUnavailable = errors.New("queue is unavailable")

const exchangeLinkEvents = "link.events"

const publishTimeout = 1 * time.Second

type RabbitmqPublisher struct {
	connection *amqp.Connection
	mu         sync.Mutex
	channel    *amqp.Channel
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

	p.mu.Lock()
	defer p.mu.Unlock()

	if p.channel == nil {
		channel, err := p.connection.Channel()
		if err != nil {
			return err
		}

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
			channel.Close()
			return err
		}

		p.channel = channel
	}

	ctx, cancel := context.WithTimeout(context.Background(), publishTimeout)
	defer cancel()

	err := p.channel.PublishWithContext(
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