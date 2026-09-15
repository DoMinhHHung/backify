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

// NewPublisher mở channel và khai báo durable topic exchange dùng để phát sự kiện nền tảng.
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

// Publish gửi payload JSON với tên sự kiện làm routing key.
// Lỗi tuần tự hóa hoặc gửi được ghi log và bỏ qua để việc phát sự kiện là best effort.
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

// Close đóng channel AMQP của publisher.
func (p *Publisher) Close() error {
	return p.channel.Close()
}
