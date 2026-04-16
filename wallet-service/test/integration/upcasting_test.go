//go:build integration

package integration_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/google/uuid"

	appaccount "github.com/savvinovan/wallet-service/internal/application/account"
	domain "github.com/savvinovan/wallet-service/internal/domain/account"
	"github.com/savvinovan/wallet-service/internal/infrastructure/eventstore"
)

var _ = Describe("Event upcasting", func() {
	ctx := context.Background()

	It("reads a v1 MoneyDeposited row as v2 with Description defaulting to empty string", func() {
		accountID := domain.NewAccountID()
		customerID := domain.NewCustomerID()

		// Open account via normal command path (writes AccountOpened with schema_version=1).
		Expect(cmdBus.Dispatch(ctx, appaccount.OpenAccountCommand{
			AccountID:  accountID,
			CustomerID: customerID,
			Currency:   "USD",
		})).To(Succeed())

		// Insert a raw v1 MoneyDeposited row directly into the DB,
		// simulating legacy data written before the v2 schema was introduced.
		// v1 shape has no "description" field.
		v1Payload := []byte(`{"amount":"100.00","currency":"USD"}`)
		_, err := pool.Exec(ctx, `
			INSERT INTO events (id, aggregate_id, aggregate_type, event_type, event_version, schema_version, payload, occurred_at)
			VALUES ($1::uuid, $2::uuid, $3, $4, 2, 1, $5, $6)
		`,
			uuid.Must(uuid.NewV7()).String(),
			string(accountID),
			domain.AggregateType,
			domain.EventTypeMoneyDeposited,
			v1Payload,
			time.Now().UTC(),
		)
		Expect(err).NotTo(HaveOccurred())

		// Load via event store — UpcasterRegistry should transparently
		// transform the v1 payload to v2 shape before deserialization.
		es := eventstore.NewPostgresEventStore(
			pool,
			eventstore.NewAccountRegistry(),
			eventstore.NewAccountUpcasterRegistry(),
		)
		loadedEvents, err := es.Load(ctx, string(accountID))
		Expect(err).NotTo(HaveOccurred())

		var deposited *domain.MoneyDeposited
		for _, e := range loadedEvents {
			if d, ok := e.(domain.MoneyDeposited); ok {
				d := d
				deposited = &d
				break
			}
		}
		Expect(deposited).NotTo(BeNil(), "MoneyDeposited event not found after load")
		Expect(deposited.Description).To(Equal(""), "v1 rows must get empty Description default via upcaster")
		Expect(deposited.Amount.Amount.StringFixed(2)).To(Equal("100.00"))
		Expect(deposited.Amount.Currency).To(Equal("USD"))
	})
})
