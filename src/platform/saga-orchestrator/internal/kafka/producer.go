package kafka

import (
	"context"
	"encoding/json"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

// Producer wraps kafka-go writer for publishing saga commands.
type Producer struct {
	writer *kafka.Writer
	log    *zap.Logger
}

func NewProducer(brokers []string, log *zap.Logger) *Producer {
	w := &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: kafka.RequireOne,
		Async:        false,
	}
	return &Producer{writer: w, log: log}
}

// Publish sends a message to the given topic with the payload JSON-encoded.
func (p *Producer) Publish(ctx context.Context, topic, key string, payload any) error {
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	msg := kafka.Message{
		Topic: topic,
		Key:   []byte(key),
		Value: b,
	}
	if err := p.writer.WriteMessages(ctx, msg); err != nil {
		p.log.Error("kafka publish failed", zap.String("topic", topic), zap.Error(err))
		return err
	}
	p.log.Debug("kafka message published", zap.String("topic", topic), zap.String("key", key))
	return nil
}

func (p *Producer) Close() error { return p.writer.Close() }
