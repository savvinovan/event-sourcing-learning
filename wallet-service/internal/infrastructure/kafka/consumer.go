package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/segmentio/kafka-go"

	"github.com/savvinovan/event-sourcing-learning/contracts/events"
	"github.com/savvinovan/event-sourcing-learning/contracts/topics"
	appaccount "github.com/savvinovan/wallet-service/internal/application/account"
	"github.com/savvinovan/wallet-service/internal/application/command"
	"github.com/savvinovan/wallet-service/config"
	domain "github.com/savvinovan/wallet-service/internal/domain/account"
)

// KYCEventConsumer consumes KYCVerified and KYCRejected events from Kafka
// and dispatches ActivateAccount / FreezeAccount commands to the wallet domain.
type KYCEventConsumer struct {
	reader *kafka.Reader
	repo   appaccount.AccountReadRepository
	bus    command.Bus
	log    *slog.Logger
}

func NewKYCEventConsumer(
	cfg *config.Config,
	repo appaccount.AccountReadRepository,
	bus command.Bus,
	log *slog.Logger,
) *KYCEventConsumer {
	addrs := strings.Split(cfg.Kafka.Brokers, ",")
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     addrs,
		GroupID:     "wallet-service",
		GroupTopics: []string{topics.KYCVerified, topics.KYCRejected},
		MinBytes:    1,
		MaxBytes:    10e6,
	})
	return &KYCEventConsumer{
		reader: reader,
		repo:   repo,
		bus:    bus,
		log:    log,
	}
}

// Run starts the consumer loop. Blocks until ctx is cancelled.
func (c *KYCEventConsumer) Run(ctx context.Context) {
	c.log.Info("KYC event consumer started")
	for {
		msg, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				c.log.Info("KYC event consumer stopped")
				return
			}
			c.log.Error("kafka: fetch message", "error", err)
			continue
		}

		if err := c.handle(ctx, msg); err != nil {
			// Infrastructure errors — don't commit, message will be redelivered on restart.
			c.log.Error("kafka: handle message", "topic", msg.Topic, "error", err)
			continue
		}

		if err := c.reader.CommitMessages(ctx, msg); err != nil {
			c.log.Error("kafka: commit message", "error", err)
		}
	}
}

func (c *KYCEventConsumer) handle(ctx context.Context, msg kafka.Message) error {
	switch msg.Topic {
	case topics.KYCVerified:
		return c.handleKYCVerified(ctx, msg.Value)
	case topics.KYCRejected:
		return c.handleKYCRejected(ctx, msg.Value)
	default:
		c.log.Warn("kafka: unknown topic, skipping", "topic", msg.Topic)
		return nil
	}
}

func (c *KYCEventConsumer) handleKYCVerified(ctx context.Context, payload []byte) error {
	var evt events.KYCVerified
	if err := json.Unmarshal(payload, &evt); err != nil {
		return fmt.Errorf("unmarshal KYCVerified: %w", err)
	}

	customerID := domain.CustomerID(evt.CustomerID)
	accountIDs, err := c.repo.GetAccountIDsByCustomerID(ctx, customerID)
	if err != nil {
		return fmt.Errorf("get accounts for customer %s: %w", customerID, err)
	}
	if len(accountIDs) == 0 {
		c.log.Warn("KYCVerified: no accounts found for customer", "customer_id", evt.CustomerID)
		return nil
	}

	for _, accountID := range accountIDs {
		cmd := appaccount.ActivateAccountCommand{AccountID: accountID}
		if err := c.bus.Dispatch(ctx, cmd); err != nil {
			if errors.Is(err, domain.ErrNotPending) {
				// Already activated or frozen — idempotent, skip.
				c.log.Info("KYCVerified: account not in pending state, skipping", "account_id", accountID)
				continue
			}
			return fmt.Errorf("activate account %s: %w", accountID, err)
		}
		c.log.Info("KYCVerified: account activated", "account_id", accountID, "customer_id", evt.CustomerID)
	}
	return nil
}

func (c *KYCEventConsumer) handleKYCRejected(ctx context.Context, payload []byte) error {
	var evt events.KYCRejected
	if err := json.Unmarshal(payload, &evt); err != nil {
		return fmt.Errorf("unmarshal KYCRejected: %w", err)
	}

	customerID := domain.CustomerID(evt.CustomerID)
	accountIDs, err := c.repo.GetAccountIDsByCustomerID(ctx, customerID)
	if err != nil {
		return fmt.Errorf("get accounts for customer %s: %w", customerID, err)
	}
	if len(accountIDs) == 0 {
		c.log.Warn("KYCRejected: no accounts found for customer", "customer_id", evt.CustomerID)
		return nil
	}

	for _, accountID := range accountIDs {
		cmd := appaccount.FreezeAccountCommand{AccountID: accountID, Reason: evt.Reason}
		if err := c.bus.Dispatch(ctx, cmd); err != nil {
			if errors.Is(err, domain.ErrNotPending) {
				// Already frozen or active — idempotent, skip.
				c.log.Info("KYCRejected: account not in pending state, skipping", "account_id", accountID)
				continue
			}
			return fmt.Errorf("freeze account %s: %w", accountID, err)
		}
		c.log.Info("KYCRejected: account frozen", "account_id", accountID, "customer_id", evt.CustomerID)
	}
	return nil
}

func (c *KYCEventConsumer) Close() error {
	return c.reader.Close()
}
