package rabbitmq

import (
	"context"
	"encoding/json"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"
)

const exchangeName = "platform.events"

// subscribedEvents là routing key Auth Service quan tâm trên exchange
// "platform.events" (cùng exchange control-plane publish, xem
// adapter/rabbitmq/publisher.go phía control-plane). Bước 1 chỉ khai báo
// queue + log; xử lý thật (tạo/xóa database, invalidate cache) là Bước 3 và
// Bước 6 — tách khai báo hạ tầng khỏi business logic để mỗi bước review độc lập.
var subscribedEvents = []string{
	"project.created",
	"project.config.updated",
	"project.deleted",
}

// Subscriber tiêu thụ sự kiện vòng đời project từ Control Plane qua một
// queue riêng của Auth Service, không chia sẻ queue với service khác.
type Subscriber struct {
	channel *amqp.Channel
	queue   string
}

// eventEnvelope chỉ đọc project_id để log có ngữ cảnh; payload đầy đủ được
// parse lại theo từng loại sự kiện khi Bước 3/6 hiện thực xử lý thật.
type eventEnvelope struct {
	ProjectID string `json:"project_id"`
}

// NewSubscriber mở channel, khai báo lại exchange "platform.events" (idempotent
// nếu control-plane đã khai báo trước) và một queue bền vững riêng cho Auth,
// bind vào đúng các routing key trong subscribedEvents.
func NewSubscriber(conn *amqp.Connection) (*Subscriber, error) {
	channel, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	if err := channel.ExchangeDeclare(exchangeName, "topic", true, false, false, false, nil); err != nil {
		return nil, err
	}

	q, err := channel.QueueDeclare("auth.project-events", true, false, false, false, nil)
	if err != nil {
		return nil, err
	}

	for _, routingKey := range subscribedEvents {
		if err := channel.QueueBind(q.Name, routingKey, exchangeName, false, nil); err != nil {
			return nil, err
		}
	}

	return &Subscriber{channel: channel, queue: q.Name}, nil
}

// Start tiêu thụ queue trong một goroutine cho tới khi ctx bị hủy hoặc channel
// đóng. Bước 1 chỉ log sự kiện nhận được (kèm ack ngay) để xác nhận đường dây
// event hoạt động; chưa ghi cache hay tạo/xóa database — đó là việc của
// usecase ở Bước 3/6, Start ở đây không nên biết business logic.
func (s *Subscriber) Start(ctx context.Context) error {
	msgs, err := s.channel.Consume(s.queue, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-msgs:
				if !ok {
					return
				}
				var envelope eventEnvelope
				_ = json.Unmarshal(msg.Body, &envelope)
				log.Info().
					Str("routing_key", msg.RoutingKey).
					Str("project_id", envelope.ProjectID).
					Msg("received control-plane event (not yet processed — Bước 3/6)")
				_ = msg.Ack(false)
			}
		}
	}()

	return nil
}

// Close đóng channel của subscriber; không đóng connection vì connection
// dùng chung với các adapter RabbitMQ khác trong App.
func (s *Subscriber) Close() error {
	return s.channel.Close()
}
