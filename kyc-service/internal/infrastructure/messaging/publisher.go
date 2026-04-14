package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"

	"github.com/savvinovan/event-sourcing-learning/contracts/events"
	"github.com/savvinovan/event-sourcing-learning/contracts/topics"
	appkyc "github.com/savvinovan/kyc-service/internal/application/kyc"
	domain "github.com/savvinovan/kyc-service/internal/domain/kyc"
)

// compile-time check
var _ appkyc.EventPublisher = (*KafkaPublisher)(nil)

// KafkaPublisher publishes KYC integration events to Kafka topics.
type KafkaPublisher struct {
	writer *kafka.Writer
}

// NewKafkaPublisher creates a publisher that writes to the given comma-separated broker list.
func NewKafkaPublisher(brokers string) *KafkaPublisher {
	addrs := strings.Split(brokers, ",")
	return &KafkaPublisher{
		writer: &kafka.Writer{
			Addr:         kafka.TCP(addrs...),
			Balancer:     &kafka.LeastBytes{},
			WriteTimeout: 10 * time.Second,
			ReadTimeout:  10 * time.Second,
		},
	}
}

func (p *KafkaPublisher) PublishKYCVerified(ctx context.Context, customerID domain.CustomerID) error {
	evt := events.KYCVerified{
		CustomerID: string(customerID),
		VerifiedAt: time.Now().UTC(),
	}
	payload, err := json.Marshal(evt)
	if err != nil {
		return fmt.Errorf("marshal KYCVerified: %w", err)
	}
	if err := p.writer.WriteMessages(ctx, kafka.Message{
		Topic: topics.KYCVerified,
		Value: payload,
	}); err != nil {
		return fmt.Errorf("publish KYCVerified: %w", err)
	}
	return nil
}

func (p *KafkaPublisher) PublishKYCRejected(ctx context.Context, customerID domain.CustomerID, reason string) error {
	evt := events.KYCRejected{
		CustomerID: string(customerID),
		Reason:     reason,
		RejectedAt: time.Now().UTC(),
	}
	payload, err := json.Marshal(evt)
	if err != nil {
		return fmt.Errorf("marshal KYCRejected: %w", err)
	}
	if err := p.writer.WriteMessages(ctx, kafka.Message{
		Topic: topics.KYCRejected,
		Value: payload,
	}); err != nil {
		return fmt.Errorf("publish KYCRejected: %w", err)
	}
	return nil
}

func (p *KafkaPublisher) Close() error {
	return p.writer.Close()
}
