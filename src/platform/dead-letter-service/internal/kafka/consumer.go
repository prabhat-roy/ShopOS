package kafka

import (
	"context"
	"log/slog"
	"time"

	kafkago "github.com/segmentio/kafka-go"

	"github.com/shopos/dead-letter-service/internal/domain"
)

// MessageSaver is the minimal interface the Kafka consumer needs to persist messages.
type MessageSaver interface {
	Save(msg *domain.DeadMessage) error
}

// Consumer reads messages from one or more DLQ topics and persists them via MessageSaver.
type Consumer struct {
	readers []*kafkago.Reader
	saver   MessageSaver
	logger  *slog.Logger
}

// NewConsumer creates a Consumer that subscribes to each topic in topics using
// the provided brokers and group ID.
func NewConsumer(brokers []string, groupID string, topics []string, saver MessageSaver, logger *slog.Logger) *Consumer {
	readers := make([]*kafkago.Reader, 0, len(topics))
	for _, topic := range topics {
		r := kafkago.NewReader(kafkago.ReaderConfig{
			Brokers:        brokers,
			GroupID:        groupID,
			Topic:          topic,
			MinBytes:       1,
			MaxBytes:       10 << 20, // 10 MiB
			CommitInterval: time.Second,
			// Start from the beginning when the group has no committed offset so
			// no DLQ messages are missed on first run.
			StartOffset: kafkago.FirstOffset,
		})
		readers = append(readers, r)
	}
	return &Consumer{
		readers: readers,
		saver:   saver,
		logger:  logger,
	}
}

// Start launches one goroutine per topic reader and blocks until ctx is cancelled.
// When ctx is cancelled all readers are closed and the function returns.
func (c *Consumer) Start(ctx context.Context) {
	done := make(chan struct{}, len(c.readers))

	for _, r := range c.readers {
		r := r // capture loop variable
		go func() {
			defer func() { done <- struct{}{} }()
			c.consume(ctx, r)
		}()
	}

	// Wait for all goroutines to finish.
	for range c.readers {
		<-done
	}
}

// Close closes all underlying Kafka readers. It is safe to call after Start returns.
func (c *Consumer) Close() {
	for _, r := range c.readers {
		if err := r.Close(); err != nil {
			c.logger.Error("kafka reader close", "topic", r.Config().Topic, "error", err)
		}
	}
}

// consume is the per-reader processing loop.
func (c *Consumer) consume(ctx context.Context, r *kafkago.Reader) {
	for {
		// FetchMessage does not auto-commit; we commit explicitly after processing.
		m, err := r.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				// Normal shutdown — context was cancelled.
				return
			}
			c.logger.Error("kafka fetch", "topic", r.Config().Topic, "error", err)
			// Back off briefly before retrying to avoid tight error loops.
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Second):
			}
			continue
		}

		msg := &domain.DeadMessage{
			Topic:     m.Topic,
			Partition: m.Partition,
			Offset:    m.Offset,
			Key:       string(m.Key),
			Payload:   m.Value,
			Status:    domain.StatusPending,
		}

		if err := c.saver.Save(msg); err != nil {
			c.logger.Error("save dead message", "topic", m.Topic, "offset", m.Offset, "error", err)
			// Still commit to avoid infinite redelivery of the same DLQ message.
		}

		// Always commit regardless of save outcome to avoid re-processing DLQ
		// messages that have already been stored (idempotency is handled by the
		// unique ID generated in the service layer).
		if err := r.CommitMessages(ctx, m); err != nil {
			if ctx.Err() != nil {
				return
			}
			c.logger.Error("kafka commit", "topic", m.Topic, "offset", m.Offset, "error", err)
		}
	}
}
