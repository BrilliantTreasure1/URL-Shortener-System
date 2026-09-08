package messagequeue

const RoutingKeyLinkClicked = "link.clicked"

type Rabbitmq interface {
	Publish(routingKey string, payload []byte) error
}