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

	walletappaccount "github.com/savvinovan/wallet-service/internal/application/account"
	walletcmd "github.com/savvinovan/wallet-service/internal/application/command"
	walletcfg "github.com/savvinovan/wallet-service/config"
	walletdb "github.com/savvinovan/wallet-service/db"
	walletstore "github.com/savvinovan/wallet-service/internal/infrastructure/eventstore"
	walletkafka "github.com/savvinovan/wallet-service/internal/infrastructure/kafka"
	walletprojector "github.com/savvinovan/wallet-service/internal/infrastructure/projector"
	walletreadmodel "github.com/savvinovan/wallet-service/internal/infrastructure/readmodel"

	kycapplication "github.com/savvinovan/kyc-service/internal/application/kyc"
	kyccmd "github.com/savvinovan/kyc-service/internal/application/command"
	kycstore "github.com/savvinovan/kyc-service/internal/infrastructure/eventstore"
	kycmessaging "github.com/savvinovan/kyc-service/internal/infrastructure/messaging"

	"github.com/savvinovan/event-sourcing-learning/contracts/topics"
	"github.com/savvinovan/event-sourcing-learning/tests/integration/helpers"
)

func TestCrossServiceIntegration(t *testing.T) {
	defer goleak.VerifyNone(t,
		goleak.IgnoreTopFunction("github.com/jackc/pgx/v5/pgxpool.(*Pool).backgroundHealthCheck"),
		goleak.IgnoreTopFunction("github.com/testcontainers/testcontainers-go.(*Reaper).connect.func1"),
		goleak.IgnoreTopFunction("github.com/onsi/ginkgo/v2/internal/interrupt_handler.(*InterruptHandler).registerForInterrupts.func2"),
		goleak.IgnoreTopFunction("github.com/segmentio/kafka-go.(*ConsumerGroup).Next"),
		goleak.IgnoreTopFunction("github.com/segmentio/kafka-go.(*ConsumerGroup).run"),
		goleak.IgnoreTopFunction("github.com/segmentio/kafka-go.(*conn).run"),
		goleak.IgnoreTopFunction("github.com/segmentio/kafka-go.(*connPool).discover"),
	)
	RegisterFailHandler(Fail)
	RunSpecs(t, "Cross-Service Integration Suite")
}

var (
	pool           *pgxpool.Pool
	broker         string
	walletCmdBus   *walletcmd.InMemoryCommandBus
	walletReadRepo walletappaccount.AccountReadRepository
	kycCmdBus      *kyccmd.InMemoryCommandBus
	kycPublisher   *kycmessaging.KafkaPublisher
	kycConsumer    *walletkafka.KYCEventConsumer
	suiteCtx       context.Context
	suiteCancel    context.CancelFunc
	cleanupPG      func()
	cleanupRP      func()
)

var _ = BeforeSuite(func() {
	suiteCtx, suiteCancel = context.WithCancel(context.Background())

	pool, _, cleanupPG = helpers.StartPostgres(suiteCtx, GinkgoT(), walletdb.Migrations)
	broker, cleanupRP = helpers.StartRedpanda(suiteCtx, GinkgoT())

	log := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))

	// --- wallet-service wiring ---
	walletES := walletstore.NewPostgresEventStore(pool, walletstore.NewAccountRegistry())
	walletReadRepo = walletreadmodel.NewPostgresReadModelRepository(pool)

	walletCmdBus = walletcmd.NewInMemoryBus()
	walletcmd.MustRegister(walletCmdBus, walletappaccount.NewOpenAccountHandler(walletES))
	walletcmd.MustRegister(walletCmdBus, walletappaccount.NewDepositMoneyHandler(walletES))
	walletcmd.MustRegister(walletCmdBus, walletappaccount.NewWithdrawMoneyHandler(walletES))
	walletcmd.MustRegister(walletCmdBus, walletappaccount.NewActivateAccountHandler(walletES))
	walletcmd.MustRegister(walletCmdBus, walletappaccount.NewFreezeAccountHandler(walletES))

	proj := walletprojector.NewRunner(
		pool,
		walletstore.NewAccountRegistry(),
		walletprojector.NewAccountProjector(),
		walletprojector.AccountProjectorName,
		100,
		50*time.Millisecond,
		log,
	)
	go proj.Run(suiteCtx) //nolint:errcheck

	testWalletCfg := &walletcfg.Config{}
	testWalletCfg.Kafka.Brokers = broker

	// Pre-create topics before starting the KYC consumer.
	helpers.CreateTopics(suiteCtx, GinkgoT(), broker,
		topics.KYCVerified, topics.KYCRejected,
	)

	kycConsumer = walletkafka.NewKYCEventConsumer(testWalletCfg, walletReadRepo, walletCmdBus, log)
	go kycConsumer.Run(suiteCtx)

	// --- kyc-service wiring ---
	kycES := kycstore.NewInMemory()
	kycPublisher = kycmessaging.NewKafkaPublisher(broker)

	kycCmdBus = kyccmd.NewInMemoryBus()
	kyccmd.MustRegister(kycCmdBus, kycapplication.NewSubmitKYCHandler(kycES))
	kyccmd.MustRegister(kycCmdBus, kycapplication.NewApproveKYCHandler(kycES, kycPublisher))
	kyccmd.MustRegister(kycCmdBus, kycapplication.NewRejectKYCHandler(kycES, kycPublisher))
})

var _ = AfterSuite(func() {
	suiteCancel()
	_ = kycConsumer.Close()
	_ = kycPublisher.Close()
	cleanupRP()
	cleanupPG()
})

var _ = BeforeEach(func() {
	_, err := pool.Exec(context.Background(),
		`TRUNCATE TABLE events, account_read_models, transaction_history RESTART IDENTITY CASCADE`)
	Expect(err).NotTo(HaveOccurred())
	_, err = pool.Exec(context.Background(),
		`UPDATE projector_checkpoints SET last_global_seq = 0, updated_at = now() WHERE projector_name = 'account_projector'`)
	Expect(err).NotTo(HaveOccurred())
})
