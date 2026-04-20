package kafka

import (
	"context"
	"log/slog"

	"github.com/segmentio/kafka-go"
)

// Dispatcher is the service method used to fan out webhook deliveries.
type Dispatcher interface {
	Dispatch(ctx context.Context, eventTopic string, payload []byte)
}

// Consumer reads Kafka topics and triggers webhook dispatch.
type Consumer struct {
	brokers    []string
	groupID    string
	topics     []string
	dispatcher Dispatcher
}

func NewConsumer(brokers []string, groupID string, topics []string, d Dispatcher) *Consumer {
	return &Consumer{brokers: brokers, groupID: groupID, topics: topics, dispatcher: d}
}

func (c *Consumer) Run(ctx context.Context) error {
	errc := make(chan error, len(c.topics))
	for _, topic := range c.topics {
		go func(t string) {
			errc <- c.consume(ctx, t)
		}(topic)
	}
	select {
	case <-ctx.Done():
		return nil
	case err := <-errc:
		return err
	}
}

func (c *Consumer) consume(ctx context.Context, topic string) error {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  c.brokers,
		Topic:    topic,
		GroupID:  c.groupID,
		MinBytes: 1,
		MaxBytes: 1 << 20,
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

		c.dispatcher.Dispatch(ctx, topic, msg.Value)

		if err := r.CommitMessages(ctx, msg); err != nil {
			slog.Warn("commit failed", "topic", topic, "err", err)
		}
	}
}
