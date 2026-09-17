package rabbitmq

import (
	"context"
	"encoding/json"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"

	"backify/services/auth/internal/port"
)

const exchangeName = "platform.events"

// subscribedEvents là routing key Auth Service quan tâm trên exchange
// "platform.events" (cùng exchange control-plane publish, xem
// adapter/rabbitmq/publisher.go phía control-plane).
var subscribedEvents = []string{
	"project.created",
	"project.config.updated",
	"project.deleted",
}

// Subscriber tiêu thụ sự kiện vòng đời project từ Control Plane qua một
// queue riêng của Auth Service, không chia sẻ queue với service khác.
// project.created/project.deleted điều khiển DatabaseManager thật (Bước 3);
// project.config.updated mới chỉ log — invalidate cache là việc của Bước 6.
type Subscriber struct {
	channel *amqp.Channel
	queue   string
	dbs     port.DatabaseManager
}

// eventEnvelope chỉ đọc project_id — đủ cho CreateDatabase/DropDatabase và
// cho log của project.config.updated; payload đầy đủ (entities/modules) sẽ
// được đọc lại qua gRPC GetProjectConfig khi Bước 6 cần ghi cache, không
// phải parse trực tiếp từ message.
type eventEnvelope struct {
	ProjectID string `json:"project_id"`
}

// NewSubscriber mở channel, khai báo lại exchange "platform.events" (idempotent
// nếu control-plane đã khai báo trước) và một queue bền vững riêng cho Auth,
// bind vào đúng các routing key trong subscribedEvents.
func NewSubscriber(conn *amqp.Connection, dbs port.DatabaseManager) (*Subscriber, error) {
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

	return &Subscriber{channel: channel, queue: q.Name, dbs: dbs}, nil
}

// Start tiêu thụ queue trong một goroutine cho tới khi ctx bị hủy hoặc channel
// đóng. Xử lý tuần tự từng message (không xử lý song song) — CreateDatabase/
// DropDatabase không cần chạy đồng thời ở quy mô MVP, và xử lý tuần tự tránh
// hai message của cùng một project đụng nhau (ví dụ created rồi deleted tới
// gần như cùng lúc).
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
				s.handle(ctx, msg)
			}
		}
	}()

	return nil
}

// handle xử lý một message theo routing key. Lỗi CreateDatabase/DropDatabase
// bị Nack không requeue (msg.Nack(false, false)) thay vì requeue vô hạn —
// chưa có dead-letter queue ở MVP, nên một message hỏng vĩnh viễn (ví dụ
// project_id không hợp lệ) sẽ bị bỏ sau lần thử đầu thay vì loop vô tận;
// đây là đánh đổi tạm chấp nhận được, DLQ là việc của phase sau.
func (s *Subscriber) handle(ctx context.Context, msg amqp.Delivery) {
	var envelope eventEnvelope
	if err := json.Unmarshal(msg.Body, &envelope); err != nil {
		log.Error().Err(err).Str("routing_key", msg.RoutingKey).Msg("failed to parse event payload")
		_ = msg.Nack(false, false)
		return
	}

	logger := log.With().Str("routing_key", msg.RoutingKey).Str("project_id", envelope.ProjectID).Logger()

	var err error
	switch msg.RoutingKey {
	case "project.created":
		err = s.dbs.CreateDatabase(ctx, envelope.ProjectID)
	case "project.deleted":
		err = s.dbs.DropDatabase(ctx, envelope.ProjectID)
	case "project.config.updated":
		logger.Info().Msg("received project.config.updated (cache invalidation deferred to Bước 6)")
	default:
		logger.Warn().Msg("received event with no handler")
	}

	if err != nil {
		logger.Error().Err(err).Msg("failed to process event")
		_ = msg.Nack(false, false)
		return
	}

	logger.Info().Msg("event processed")
	_ = msg.Ack(false)
}

// Close đóng channel của subscriber; không đóng connection vì connection
// dùng chung với các adapter RabbitMQ khác trong App.
func (s *Subscriber) Close() error {
	return s.channel.Close()
}
