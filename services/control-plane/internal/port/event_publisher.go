package port

import "context"

type Event struct {
	Name    string
	Payload interface{}
}

type EventPublisher interface {
	Publish(ctx context.Context, event Event) error
}
