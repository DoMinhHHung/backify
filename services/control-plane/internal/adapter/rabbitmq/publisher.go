package rabbitmq

import (
	"context"
	"encoding/json"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"

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
		log.Error().Err(err).Str("event", event.Name).Msg("failed to marshal event payload, dropping event")
		return nil
	}

	if err := p.channel.PublishWithContext(ctx, exchangeName, event.Name, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        body,
	}); err != nil {
		log.Error().Err(err).Str("event", event.Name).Msg("failed to publish event, dropping event")
		return nil
	}

	return nil
}

func (p *Publisher) Close() error {
	return p.channel.Close()
}
