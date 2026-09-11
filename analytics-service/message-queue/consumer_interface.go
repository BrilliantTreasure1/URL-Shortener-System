package messagequeue

import (
	"context"
	"errors"
)

const RoutingKeyLinkClicked = "link.clicked"

const ExchangeLinkEvents = "link.events"

const QueueLinkClicks = "link.clicks"

var ErrQueueUnavailable = errors.New("queue is unavailable")

type MessageHandler func(payload []byte) error

type Consumer interface {
	Consume(ctx context.Context, handler MessageHandler) error
}