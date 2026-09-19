package rabbitmq

import (
	"context"
	"encoding/json"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"

	"backify/services/auth/internal/port"
)

// Publisher implement port.EventPublisher, publish lên cùng exchange
// platform.events mà Subscriber đã bind. Best-effort: lỗi marshal/publish
// chỉ log, luôn trả nil để usecase không bị fail vì side-effect.
type Publisher struct {
	channel *amqp.Channel
}

// NewPublisher mở channel và khai báo topic exchange bền vững (idempotent
// với control-plane publisher và auth subscriber).
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
		_ = channel.Close()
		return nil, err
	}

	return &Publisher{channel: channel}, nil
}

// Publish gửi payload JSON với routing key = event.Name. Lỗi được log và bỏ qua.
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

// Close đóng channel của publisher.
func (p *Publisher) Close() error {
	return p.channel.Close()
}

var _ port.EventPublisher = (*Publisher)(nil)
