package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/DoMinhHHung/backify/services/runtime/internal/port"
)

const exchangeName = "backify.events"

type RabbitConsumer struct {
	url     string
	queue   string
	handler port.ConfigEventHandler
	conn    *amqp.Connection
	channel *amqp.Channel
}

func NewRabbitConsumer(url, queue string, handler port.ConfigEventHandler) *RabbitConsumer {
	return &RabbitConsumer{url: url, queue: queue, handler: handler}
}

type eventPayload struct {
	Event      string `json:"event"`
	ProjectID  string `json:"projectId"`
	Slug       string `json:"slug"`
	SchemaName string `json:"schemaName"`
	OccurredAt string `json:"occurredAt"`
}

func (c *RabbitConsumer) Start(ctx context.Context) error {
	conn, err := amqp.Dial(c.url)
	if err != nil {
		return fmt.Errorf("rabbit dial: %w", err)
	}
	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return fmt.Errorf("rabbit channel: %w", err)
	}
	c.conn = conn
	c.channel = ch

	if err := ch.ExchangeDeclare(exchangeName, "topic", true, false, false, false, nil); err != nil {
		return fmt.Errorf("declare exchange: %w", err)
	}

	q, err := ch.QueueDeclare(c.queue, true, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("declare queue: %w", err)
	}

	for _, key := range []string{"project.created", "project.config.updated"} {
		if err := ch.QueueBind(q.Name, key, exchangeName, false, nil); err != nil {
			return fmt.Errorf("bind %s: %w", key, err)
		}
	}

	if err := ch.Qos(10, 0, false); err != nil {
		return err
	}

	deliveries, err := ch.Consume(q.Name, "runtime", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("consume: %w", err)
	}

	log.Printf("rabbit consumer started queue=%s", c.queue)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case d, ok := <-deliveries:
				if !ok {
					return
				}
				c.handleDelivery(ctx, d)
			}
		}
	}()

	return nil
}

func (c *RabbitConsumer) handleDelivery(ctx context.Context, d amqp.Delivery) {
	var payload eventPayload
	if err := json.Unmarshal(d.Body, &payload); err != nil {
		log.Printf("rabbit invalid payload: %v", err)
		_ = d.Nack(false, false)
		return
	}
	if payload.ProjectID == "" {
		_ = d.Nack(false, false)
		return
	}

	hctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var err error
	switch payload.Event {
	case "project.created":
		err = c.handler.HandleProjectCreated(hctx, payload.ProjectID)
	case "project.config.updated":
		err = c.handler.HandleConfigUpdated(hctx, payload.ProjectID)
	default:
		_ = d.Ack(false)
		return
	}

	if err != nil {
		log.Printf("rabbit handle error event=%s project=%s: %v", payload.Event, payload.ProjectID, err)
		_ = d.Nack(false, true)
		return
	}
	_ = d.Ack(false)
}

func (c *RabbitConsumer) Close() {
	if c.channel != nil {
		_ = c.channel.Close()
	}
	if c.conn != nil {
		_ = c.conn.Close()
	}
}
