//go:build integration

package integration_test

import (
	"context"
	"embed"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	appkyc "github.com/savvinovan/kyc-service/internal/application/kyc"
	domain "github.com/savvinovan/kyc-service/internal/domain/kyc"
	"github.com/savvinovan/event-sourcing-learning/contracts/events"
	"github.com/savvinovan/event-sourcing-learning/contracts/topics"
	"github.com/savvinovan/kyc-service/test/integration/helpers"
)

//go:embed testdata/fixtures/approve_kyc.yaml
var approveKYCFS embed.FS

//go:embed testdata/fixtures/reject_kyc.yaml
var rejectKYCFS embed.FS

type approveKYCInput struct {
	CustomerIDSuffix string `yaml:"customer_id_suffix"`
}

type rejectKYCInput struct {
	CustomerIDSuffix string `yaml:"customer_id_suffix"`
	Reason           string `yaml:"reason"`
}

var _ = Describe("KYC Kafka publisher", func() {
	ctx := context.Background()

	Describe("Approving KYC", func() {
		fixtures := helpers.LoadFixtures[approveKYCInput](approveKYCFS, "testdata/fixtures/approve_kyc.yaml")

		for _, f := range fixtures {
			f := f
			It(f.Scenario, func() {
				verificationID := domain.NewVerificationID()
				customerID := domain.CustomerID(f.Input.CustomerIDSuffix)

				// Subscribe to topic BEFORE the action to capture the message.
				consumer := helpers.NewTestKafkaConsumer(ctx, GinkgoT(), broker, topics.KYCVerified)
				defer consumer.Close()

				Expect(cmdBus.Dispatch(ctx, appkyc.SubmitKYCCommand{
					VerificationID: verificationID,
					CustomerID:     customerID,
				})).To(Succeed())
				Expect(cmdBus.Dispatch(ctx, appkyc.ApproveKYCCommand{
					VerificationID: verificationID,
				})).To(Succeed())

				var msg events.KYCVerified
				consumer.ReadMessageInto(GinkgoT(), kafkaTimeout, &msg)
				Expect(msg.CustomerID).To(Equal(f.Input.CustomerIDSuffix))
				Expect(msg.VerifiedAt.IsZero()).To(BeFalse())
			})
		}
	})

	Describe("Rejecting KYC", func() {
		fixtures := helpers.LoadFixtures[rejectKYCInput](rejectKYCFS, "testdata/fixtures/reject_kyc.yaml")

		for _, f := range fixtures {
			f := f
			It(f.Scenario, func() {
				verificationID := domain.NewVerificationID()
				customerID := domain.CustomerID(f.Input.CustomerIDSuffix)

				// Subscribe to topic BEFORE the action.
				consumer := helpers.NewTestKafkaConsumer(ctx, GinkgoT(), broker, topics.KYCRejected)
				defer consumer.Close()

				Expect(cmdBus.Dispatch(ctx, appkyc.SubmitKYCCommand{
					VerificationID: verificationID,
					CustomerID:     customerID,
				})).To(Succeed())
				Expect(cmdBus.Dispatch(ctx, appkyc.RejectKYCCommand{
					VerificationID: verificationID,
					Reason:         f.Input.Reason,
				})).To(Succeed())

				var msg events.KYCRejected
				consumer.ReadMessageInto(GinkgoT(), kafkaTimeout, &msg)
				Expect(msg.CustomerID).To(Equal(f.Input.CustomerIDSuffix))
				Expect(msg.Reason).To(Equal(f.Input.Reason))
				Expect(msg.RejectedAt.IsZero()).To(BeFalse())
			})
		}
	})
})

const kafkaTimeout = 15 * helpers.Second
