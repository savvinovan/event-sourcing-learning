//go:build integration

package integration_test

import (
	"context"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/goleak"

	appkyc "github.com/savvinovan/kyc-service/internal/application/kyc"
	"github.com/savvinovan/kyc-service/internal/application/command"
	"github.com/savvinovan/kyc-service/internal/application/query"
	"github.com/savvinovan/kyc-service/internal/infrastructure/eventstore"
	kycmessaging "github.com/savvinovan/kyc-service/internal/infrastructure/messaging"
	"github.com/savvinovan/event-sourcing-learning/contracts/topics"
	"github.com/savvinovan/kyc-service/test/integration/helpers"
)

func TestIntegration(t *testing.T) {
	defer goleak.VerifyNone(t,
		goleak.IgnoreTopFunction("github.com/testcontainers/testcontainers-go.(*Reaper).connect.func1"),
		goleak.IgnoreTopFunction("github.com/onsi/ginkgo/v2/internal/interrupt_handler.(*InterruptHandler).registerForInterrupts.func2"),
		goleak.IgnoreTopFunction("github.com/segmentio/kafka-go.(*conn).run"),
		goleak.IgnoreTopFunction("github.com/segmentio/kafka-go.(*connPool).discover"),
	)
	RegisterFailHandler(Fail)
	RunSpecs(t, "KYC Service Integration Suite")
}

var (
	broker      string
	publisher   *kycmessaging.KafkaPublisher
	cmdBus      *command.InMemoryCommandBus
	qryBus      *query.InMemoryQueryBus
	suiteCtx    context.Context
	suiteCancel context.CancelFunc
	cleanupRP   func()
)

var _ = BeforeSuite(func() {
	suiteCtx, suiteCancel = context.WithCancel(context.Background())

	broker, cleanupRP = helpers.StartRedpanda(suiteCtx, GinkgoT())

	// Pre-create topics so kafka.DialLeader in test consumers doesn't hang.
	helpers.CreateTopics(suiteCtx, GinkgoT(), broker,
		topics.KYCVerified, topics.KYCRejected,
	)

	// Wire application layer.
	es := eventstore.NewInMemory()
	publisher = kycmessaging.NewKafkaPublisher(broker)

	cmdBus = command.NewInMemoryBus()
	command.MustRegister(cmdBus, appkyc.NewSubmitKYCHandler(es))
	command.MustRegister(cmdBus, appkyc.NewApproveKYCHandler(es, publisher))
	command.MustRegister(cmdBus, appkyc.NewRejectKYCHandler(es, publisher))

	qryBus = query.NewInMemoryBus()
	query.MustRegister(qryBus, appkyc.NewGetKYCStatusHandler(es))
})

var _ = AfterSuite(func() {
	suiteCancel()
	_ = publisher.Close()
	cleanupRP()
})
