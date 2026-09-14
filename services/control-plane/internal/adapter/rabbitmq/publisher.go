package rabbitmq

import (
	"context"
	"encoding/json"

	amqp "github.com/rabbitmq/amqp091-go"

	"backify/services/control-plane/internal/port"
)

const exchangeName = "platform.events"

type Publisher struct {
	channel *amqp.Channel
}

func NewPublisher(conn *amqp.Connection) (*Publisher, error) {
	channel, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	if err := channel.ExchangeDeclare(
		exchangeName,
		"topic",
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return nil, err
	}

	return &Publisher{channel: channel}, nil
}

func (p *Publisher) Publish(ctx context.Context, event port.Event) error {
	body, err := json.Marshal(event.Payload)
	if err != nil {
		return err
	}

	return p.channel.PublishWithContext(ctx, exchangeName, event.Name, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        body,
	})
}

func (p *Publisher) Close() error {
	return p.channel.Close()
}
