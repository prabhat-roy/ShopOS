// Package consumer implements a Kafka consumer for the push.send topic.
package consumer

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"time"

	"github.com/segmentio/kafka-go"

	"github.com/shopos/push-notification-service/internal/config"
	"github.com/shopos/push-notification-service/internal/domain"
	"github.com/shopos/push-notification-service/internal/sender"
	"github.com/shopos/push-notification-service/internal/store"
)

// Consumer reads push.send messages from Kafka and dispatches them for delivery.
type Consumer struct {
	reader  *kafka.Reader
	sender  *sender.Sender
	store   *store.Store
	running bool
}

// New creates a Consumer wired to the given config, sender and store.
func New(cfg *config.Config, s *sender.Sender, st *store.Store) *Consumer {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        cfg.KafkaBrokers,
		Topic:          cfg.KafkaTopic,
		GroupID:        cfg.KafkaGroupID,
		MinBytes:       1,
		MaxBytes:       1 << 20, // 1 MiB
		CommitInterval: 0,       // manual commit
		StartOffset:    kafka.FirstOffset,
		// Back-off on fetch errors
		ReadBackoffMin: 500 * time.Millisecond,
		ReadBackoffMax: 10 * time.Second,
	})

	return &Consumer{
		reader: r,
		sender: s,
		store:  st,
	}
}

// Run starts the consume loop and blocks until ctx is cancelled or an
// unrecoverable error occurs.
func (c *Consumer) Run(ctx context.Context) {
	c.running = true
	log.Printf("[consumer] Starting — topic=%s group=%s brokers=%v",
		c.reader.Config().Topic,
		c.reader.Config().GroupID,
		c.reader.Config().Brokers,
	)

	for {
		select {
		case <-ctx.Done():
			log.Printf("[consumer] Context cancelled — shutting down")
			c.running = false
			if err := c.reader.Close(); err != nil {
				log.Printf("[consumer] Error closing reader: %v", err)
			}
			return
		default:
		}

		msg, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, io.EOF) {
				log.Printf("[consumer] Fetch stopped: %v", err)
				c.running = false
				return
			}
			log.Printf("[consumer] FetchMessage error: %v — retrying", err)
			time.Sleep(2 * time.Second)
			continue
		}

		if err := c.handleMessage(ctx, msg); err != nil {
			// handleMessage has already logged the detail; do NOT commit so the
			// message is reprocessed on the next startup.
			log.Printf("[consumer] Skipping commit for offset=%d due to error: %v", msg.Offset, err)
			continue
		}

		// Commit only after successful processing and persistence.
		if err := c.reader.CommitMessages(ctx, msg); err != nil {
			log.Printf("[consumer] CommitMessages failed for offset=%d: %v", msg.Offset, err)
		}
	}
}

// IsRunning reports whether the consume loop is active.
func (c *Consumer) IsRunning() bool {
	return c.running
}

// -------------------------------------------------------------------
// Internal
// -------------------------------------------------------------------

func (c *Consumer) handleMessage(ctx context.Context, msg kafka.Message) error {
	logCtx := log.New(log.Writer(), "", log.LstdFlags)

	// Step 1: JSON decode
	var pushMsg domain.PushMessage
	if err := json.Unmarshal(msg.Value, &pushMsg); err != nil {
		logCtx.Printf("[consumer] JSON decode error at offset=%d: %v — skipping", msg.Offset, err)
		// Return nil so the offset is committed and we don't spin on a poison pill.
		return nil
	}

	// Step 2: Basic presence check
	if pushMsg.MessageID == "" {
		logCtx.Printf("[consumer] Missing messageId at offset=%d — skipping", msg.Offset)
		return nil
	}

	logCtx.Printf("[consumer] Processing messageId=%s platform=%s offset=%d",
		pushMsg.MessageID, pushMsg.Platform, msg.Offset)

	// Step 3: Send
	record := c.sender.Send(pushMsg)

	// Step 4: Persist
	c.store.Save(record)

	logCtx.Printf("[consumer] Processed messageId=%s status=%s", record.MessageID, record.Status)
	return nil
}
