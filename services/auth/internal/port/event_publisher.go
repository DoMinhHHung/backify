package port

import "context"

// Event là envelope sự kiện nội bộ — lặp lại đúng shape của control-plane,
// không import chéo (hai Go module riêng).
type Event struct {
	Name    string
	Payload interface{}
}

// EventPublisher phát sự kiện lên message bus. Triển khai RabbitMQ là
// best-effort: lỗi publish chỉ log, không trả lỗi lên usecase.
type EventPublisher interface {
	Publish(ctx context.Context, event Event) error
}
