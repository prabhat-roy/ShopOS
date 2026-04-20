package kafka

import (
	"context"
	"log/slog"

	"github.com/segmentio/kafka-go"
)

// Handler processes a single Kafka message payload.
type Handler func(ctx context.Context, msg []byte) error

// Consumer reads from multiple Kafka topics concurrently.
type Consumer struct {
	brokers  []string
	groupID  string
	handlers map[string]Handler
}

func NewConsumer(brokers []string, groupID string) *Consumer {
	return &Consumer{
		brokers:  brokers,
		groupID:  groupID,
		handlers: make(map[string]Handler),
	}
}

func (c *Consumer) Register(topic string, h Handler) {
	c.handlers[topic] = h
}

// Run starts one goroutine per registered topic and blocks until ctx is done.
func (c *Consumer) Run(ctx context.Context) error {
	errc := make(chan error, len(c.handlers))

	for topic, h := range c.handlers {
		go func(topic string, h Handler) {
			errc <- c.consume(ctx, topic, h)
		}(topic, h)
	}

	select {
	case <-ctx.Done():
		return nil
	case err := <-errc:
		return err
	}
}

func (c *Consumer) consume(ctx context.Context, topic string, h Handler) error {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  c.brokers,
		Topic:    topic,
		GroupID:  c.groupID,
		MinBytes: 1,
		MaxBytes: 1 << 20, // 1 MB
	})
	defer r.Close()

	for {
		msg, err := r.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return err
		}

		if err := h(ctx, msg.Value); err != nil {
			slog.Error("handler error", "topic", topic, "err", err)
			// do not commit — message will be redelivered
			continue
		}

		if err := r.CommitMessages(ctx, msg); err != nil {
			slog.Warn("commit failed", "topic", topic, "err", err)
		}
	}
}
