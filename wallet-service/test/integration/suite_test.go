//go:build integration

package integration_test

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/goleak"

	appaccount "github.com/savvinovan/wallet-service/internal/application/account"
	"github.com/savvinovan/wallet-service/internal/application/command"
	"github.com/savvinovan/wallet-service/config"
	"github.com/savvinovan/wallet-service/db"
	"github.com/savvinovan/wallet-service/internal/infrastructure/eventstore"
	"github.com/savvinovan/wallet-service/internal/infrastructure/kafka"
	"github.com/savvinovan/wallet-service/internal/infrastructure/projector"
	"github.com/savvinovan/wallet-service/internal/infrastructure/readmodel"
	"github.com/savvinovan/event-sourcing-learning/contracts/topics"
	"github.com/savvinovan/wallet-service/test/integration/helpers"
)

func TestIntegration(t *testing.T) {
	defer goleak.VerifyNone(t,
		// pgxpool background health-checker
		goleak.IgnoreTopFunction("github.com/jackc/pgx/v5/pgxpool.(*Pool).backgroundHealthCheck"),
		// testcontainers Reaper goroutines
		goleak.IgnoreTopFunction("github.com/testcontainers/testcontainers-go.(*Reaper).connect.func1"),
		// Ginkgo interrupt handler (internal, always present during suite run)
		goleak.IgnoreTopFunction("github.com/onsi/ginkgo/v2/internal/interrupt_handler.(*InterruptHandler).registerForInterrupts.func2"),
		// kafka-go consumer group goroutines (torn down asynchronously by Close)
		goleak.IgnoreTopFunction("github.com/segmentio/kafka-go.(*ConsumerGroup).Next"),
		goleak.IgnoreTopFunction("github.com/segmentio/kafka-go.(*ConsumerGroup).run"),
		goleak.IgnoreTopFunction("github.com/segmentio/kafka-go.(*conn).run"),
		goleak.IgnoreTopFunction("github.com/segmentio/kafka-go.(*connPool).discover"),
	)
	RegisterFailHandler(Fail)
	RunSpecs(t, "Wallet Service Integration Suite")
}

var (
	pool        *pgxpool.Pool
	broker      string
	cmdBus      *command.InMemoryCommandBus
	readRepo    appaccount.AccountReadRepository
	producer    *helpers.TestKafkaProducer
	kycConsumer *kafka.KYCEventConsumer
	suiteCtx    context.Context
	suiteCancel context.CancelFunc
	cleanupPG   func()
	cleanupRP   func()
)

var _ = BeforeSuite(func() {
	suiteCtx, suiteCancel = context.WithCancel(context.Background())

	pool, _, cleanupPG = helpers.StartPostgres(suiteCtx, GinkgoT(), db.Migrations)
	broker, cleanupRP = helpers.StartRedpanda(suiteCtx, GinkgoT())

	// Wire application layer.
	es := eventstore.NewPostgresEventStore(pool, eventstore.NewAccountRegistry())
	readRepo = readmodel.NewPostgresReadModelRepository(pool)

	cmdBus = command.NewInMemoryBus()
	command.MustRegister(cmdBus, appaccount.NewOpenAccountHandler(es))
	command.MustRegister(cmdBus, appaccount.NewDepositMoneyHandler(es))
	command.MustRegister(cmdBus, appaccount.NewWithdrawMoneyHandler(es))
	command.MustRegister(cmdBus, appaccount.NewActivateAccountHandler(es))
	command.MustRegister(cmdBus, appaccount.NewFreezeAccountHandler(es))

	log := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))

	// Projector tails events and updates read models; short poll for fast test feedback.
	proj := projector.NewRunner(
		pool,
		eventstore.NewAccountRegistry(),
		projector.NewAccountProjector(),
		projector.AccountProjectorName,
		100,
		50*time.Millisecond,
		log,
	)
	go proj.Run(suiteCtx) //nolint:errcheck

	// KYC event consumer: translates KYCVerified/KYCRejected Kafka messages into commands.
	testCfg := &config.Config{}
	testCfg.Kafka.Brokers = broker
	kycConsumer = kafka.NewKYCEventConsumer(testCfg, readRepo, cmdBus, log)
	go kycConsumer.Run(suiteCtx)

	// Pre-create Kafka topics — Redpanda dev mode doesn't auto-create on first write.
	helpers.CreateTopics(suiteCtx, GinkgoT(), broker,
		topics.KYCVerified, topics.KYCRejected,
	)

	producer = helpers.NewTestKafkaProducer(broker)
})

var _ = AfterSuite(func() {
	suiteCancel() // stops projector + KYC consumer Run loops
	_ = kycConsumer.Close()
	_ = producer.Close()
	cleanupRP()
	cleanupPG()
})

var _ = BeforeEach(func() {
	helpers.TruncateAll(context.Background(), GinkgoT(), pool)
})
