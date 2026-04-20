package kafka

import (
	"context"
	"encoding/json"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

// Handler is called for each consumed message.
type Handler func(ctx context.Context, topic, key string, payload []byte) error

// Consumer reads from one or more Kafka topics and dispatches to registered handlers.
type Consumer struct {
	readers  []*kafka.Reader
	handlers map[string]Handler
	log      *zap.Logger
}

func NewConsumer(brokers []string, groupID string, topics []string, log *zap.Logger) *Consumer {
	readers := make([]*kafka.Reader, 0, len(topics))
	for _, topic := range topics {
		r := kafka.NewReader(kafka.ReaderConfig{
			Brokers:  brokers,
			GroupID:  groupID,
			Topic:    topic,
			MinBytes: 1,
			MaxBytes: 1e6,
		})
		readers = append(readers, r)
	}
	return &Consumer{
		readers:  readers,
		handlers: make(map[string]Handler),
		log:      log,
	}
}

// Register maps a topic to a handler function.
func (c *Consumer) Register(topic string, h Handler) {
	c.handlers[topic] = h
}

// Start launches one goroutine per reader. Blocks until ctx is cancelled.
func (c *Consumer) Start(ctx context.Context) {
	for _, r := range c.readers {
		go c.consume(ctx, r)
	}
}

func (c *Consumer) consume(ctx context.Context, r *kafka.Reader) {
	for {
		msg, err := r.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return // context cancelled — clean shutdown
			}
			c.log.Error("kafka fetch error", zap.Error(err))
			continue
		}

		h, ok := c.handlers[msg.Topic]
		if !ok {
			c.log.Warn("no handler for topic", zap.String("topic", msg.Topic))
			r.CommitMessages(ctx, msg)
			continue
		}

		if err := h(ctx, msg.Topic, string(msg.Key), msg.Value); err != nil {
			c.log.Error("message handler error",
				zap.String("topic", msg.Topic),
				zap.Error(err),
			)
			// Don't commit on error — message will be redelivered
			continue
		}
		r.CommitMessages(ctx, msg)
	}
}

func (c *Consumer) Close() {
	for _, r := range c.readers {
		r.Close()
	}
}

// ParseJSON is a helper for handlers to unmarshal message payloads.
func ParseJSON(data []byte, v any) error {
	return json.Unmarshal(data, v)
}
