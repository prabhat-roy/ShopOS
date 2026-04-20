package scheduler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	"github.com/shopos/digest-service/internal/domain"
	"github.com/shopos/digest-service/internal/store"
)

// emailPayload is the Kafka message body sent to the email-service.
type emailPayload struct {
	To      string            `json:"to"`
	Subject string            `json:"subject"`
	Body    string            `json:"body"`
	Meta    map[string]string `json:"meta"`
}

// DigestScheduler periodically checks for due digest configs and dispatches emails.
type DigestScheduler struct {
	store    store.Storer
	writer   *kafka.Writer
	interval time.Duration
	topic    string
}

// New creates a DigestScheduler.
func New(s store.Storer, writer *kafka.Writer, topic string, interval time.Duration) *DigestScheduler {
	return &DigestScheduler{
		store:    s,
		writer:   writer,
		interval: interval,
		topic:    topic,
	}
}

// Start runs the scheduler loop until ctx is cancelled.
func (d *DigestScheduler) Start(ctx context.Context) {
	log.Printf("digest scheduler started, check interval=%s", d.interval)
	ticker := time.NewTicker(d.interval)
	defer ticker.Stop()

	// Run once immediately on startup so we don't wait a full interval.
	d.checkAndSendDue(ctx)

	for {
		select {
		case <-ctx.Done():
			log.Println("digest scheduler stopping")
			return
		case <-ticker.C:
			d.checkAndSendDue(ctx)
		}
	}
}

// checkAndSendDue finds all ACTIVE configs that are due and dispatches digest emails.
func (d *DigestScheduler) checkAndSendDue(ctx context.Context) {
	now := time.Now().UTC()
	configs, err := d.store.ListDueConfigs(ctx, now)
	if err != nil {
		log.Printf("scheduler: ListDueConfigs error: %v", err)
		return
	}
	if len(configs) == 0 {
		return
	}
	log.Printf("scheduler: %d digest(s) due", len(configs))

	for _, cfg := range configs {
		if err := d.sendDigest(ctx, cfg, now); err != nil {
			log.Printf("scheduler: sendDigest failed for config %s: %v", cfg.ID, err)
		}
	}
}

// sendDigest builds a digest payload, publishes it to Kafka, records the run,
// and advances the config's next_send_at.
func (d *DigestScheduler) sendDigest(ctx context.Context, cfg domain.DigestConfig, now time.Time) error {
	// Build mocked digest content — in production this would aggregate
	// recent orders, promotions etc. from other services via gRPC/Kafka.
	itemCount := 5 // mocked
	subject := fmt.Sprintf("Your %s digest from ShopOS", cfg.Frequency)
	body := fmt.Sprintf(
		"Hello,\n\nHere is your %s digest summary.\n\n• %d new activity item(s) since your last digest.\n\nVisit ShopOS to see the full details.\n\nBest regards,\nThe ShopOS Team",
		cfg.Frequency, itemCount,
	)

	payload := emailPayload{
		To:      cfg.Email,
		Subject: subject,
		Body:    body,
		Meta: map[string]string{
			"digest_config_id": cfg.ID.String(),
			"user_id":          cfg.UserID.String(),
			"frequency":        string(cfg.Frequency),
		},
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal email payload: %w", err)
	}

	msg := kafka.Message{
		Topic: d.topic,
		Key:   []byte(cfg.UserID.String()),
		Value: data,
		Time:  now,
	}
	if err := d.writer.WriteMessages(ctx, msg); err != nil {
		// Record failed run.
		run := domain.DigestRun{
			ID:        uuid.New(),
			ConfigID:  cfg.ID,
			SentAt:    now,
			ItemCount: 0,
			Status:    "FAILED",
			ErrorMsg:  err.Error(),
		}
		_ = d.store.SaveRun(ctx, run)
		return fmt.Errorf("kafka WriteMessages: %w", err)
	}

	// Record successful run.
	run := domain.DigestRun{
		ID:        uuid.New(),
		ConfigID:  cfg.ID,
		SentAt:    now,
		ItemCount: itemCount,
		Status:    "SUCCESS",
	}
	if err := d.store.SaveRun(ctx, run); err != nil {
		log.Printf("scheduler: SaveRun error for config %s: %v", cfg.ID, err)
	}

	// Update last_sent_at and advance next_send_at.
	if err := d.store.UpdateLastSent(ctx, cfg.ID, now); err != nil {
		log.Printf("scheduler: UpdateLastSent error for config %s: %v", cfg.ID, err)
	}
	next := ComputeNextSend(cfg.Frequency, now)
	if err := d.store.UpdateNextSend(ctx, cfg.ID, next); err != nil {
		log.Printf("scheduler: UpdateNextSend error for config %s: %v", cfg.ID, err)
	}
	log.Printf("scheduler: digest sent for config %s (user=%s), next=%s", cfg.ID, cfg.UserID, next.Format(time.RFC3339))
	return nil
}

// ComputeNextSend returns the next scheduled time based on frequency.
func ComputeNextSend(frequency domain.DigestFrequency, from time.Time) time.Time {
	switch frequency {
	case domain.FrequencyWeekly:
		return from.Add(7 * 24 * time.Hour)
	default: // DAILY
		return from.Add(24 * time.Hour)
	}
}
