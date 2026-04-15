//go:build integration

package helpers

import (
	"context"
	"encoding/json"
	"time"

	"github.com/segmentio/kafka-go"

	"github.com/savvinovan/event-sourcing-learning/contracts/events"
	"github.com/savvinovan/event-sourcing-learning/contracts/topics"
)

// TestKafkaProducer publishes integration events to Kafka for use in tests.
// Use it to simulate upstream services (e.g. kyc-service) in wallet-service tests.
type TestKafkaProducer struct {
	writer *kafka.Writer
}

func NewTestKafkaProducer(broker string) *TestKafkaProducer {
	return &TestKafkaProducer{
		writer: &kafka.Writer{
			Addr:                   kafka.TCP(broker),
			Balancer:               &kafka.LeastBytes{},
			WriteTimeout:           10 * time.Second,
			AllowAutoTopicCreation: true,
		},
	}
}

// PublishKYCVerified writes a KYCVerified event for the given customerID to Kafka.
func (p *TestKafkaProducer) PublishKYCVerified(ctx context.Context, t TB, customerID string) {
	t.Helper()
	evt := events.KYCVerified{
		CustomerID: customerID,
		VerifiedAt: time.Now().UTC(),
	}
	p.publish(ctx, t, topics.KYCVerified, evt)
}

// PublishKYCRejected writes a KYCRejected event for the given customerID to Kafka.
func (p *TestKafkaProducer) PublishKYCRejected(ctx context.Context, t TB, customerID, reason string) {
	t.Helper()
	evt := events.KYCRejected{
		CustomerID: customerID,
		Reason:     reason,
		RejectedAt: time.Now().UTC(),
	}
	p.publish(ctx, t, topics.KYCRejected, evt)
}

func (p *TestKafkaProducer) publish(ctx context.Context, t TB, topic string, payload any) {
	t.Helper()
	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal kafka message for topic %s: %v", topic, err)
	}
	if err := p.writer.WriteMessages(ctx, kafka.Message{
		Topic: topic,
		Value: data,
	}); err != nil {
		t.Fatalf("write kafka message to %s: %v", topic, err)
	}
}

func (p *TestKafkaProducer) Close() error {
	return p.writer.Close()
}
