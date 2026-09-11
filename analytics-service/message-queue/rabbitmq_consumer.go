package messagequeue

import (
	"context"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitmqConsumer struct {
	connection *amqp.Connection
}

func NewRabbitmqConsumer(connection *amqp.Connection) *RabbitmqConsumer {
	return &RabbitmqConsumer{
		connection: connection,
	}
}

func (c *RabbitmqConsumer) Consume(ctx context.Context, handler MessageHandler) error {

	if c.connection == nil {
		return ErrQueueUnavailable
	}

	channel, err := c.connection.Channel()
	if err != nil {
		return err
	}
	defer channel.Close()

	err = channel.ExchangeDeclare(
		ExchangeLinkEvents,
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

	queue, err := channel.QueueDeclare(
		QueueLinkClicks,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	err = channel.QueueBind(
		queue.Name,
		RoutingKeyLinkClicked,
		ExchangeLinkEvents,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	deliveries, err := channel.Consume(
		queue.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return nil

		case delivery, ok := <-deliveries:
			if !ok {
				return nil
			}

			if err := handler(delivery.Body); err != nil {
				_ = delivery.Nack(false, true)
				continue
			}

			_ = delivery.Ack(false)
		}
	}
}

var _ Consumer = (*RabbitmqConsumer)(nil)