//go:build integration

package integration_test

import (
	"context"
	"embed"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	appaccount "github.com/savvinovan/wallet-service/internal/application/account"
	domain "github.com/savvinovan/wallet-service/internal/domain/account"
	"github.com/savvinovan/wallet-service/test/integration/helpers"
)

//go:embed testdata/fixtures/kyc_rejected.yaml
var kycRejectedFS embed.FS

//go:embed testdata/fixtures/kyc_verified.yaml
var kycVerifiedFS embed.FS

var _ = Describe("KYC Kafka consumer", func() {
	ctx := context.Background()
	fixtures := helpers.LoadFixtures[kycConsumerInput](kycVerifiedFS, "testdata/fixtures/kyc_verified.yaml")

	for _, f := range fixtures {
		f := f
		It(f.Scenario, func() {
			accountID := domain.NewAccountID()
			customerID := domain.NewCustomerID()

			// Open account (Pending state).
			Expect(cmdBus.Dispatch(ctx, appaccount.OpenAccountCommand{
				AccountID:  accountID,
				CustomerID: customerID,
				Currency:   f.Input.Currency,
			})).To(Succeed())

			// Wait for read model to reflect Pending before publishing KYC event.
			Eventually(func(g Gomega) {
				result, err := readRepo.GetBalance(ctx, accountID)
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(result.Status).To(Equal("Pending"))
			}).WithTimeout(projectorTimeout).WithPolling(projectorPoll).Should(Succeed())

			// Publish KYCVerified — the KYC consumer will dispatch ActivateAccountCommand.
			producer.PublishKYCVerified(ctx, GinkgoT(), string(customerID))

			// Assert: Eventually the account becomes Active via Kafka → command → event → projector.
			Eventually(func(g Gomega) {
				result, err := readRepo.GetBalance(ctx, accountID)
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(result.Status).To(Equal("Active"))
			}).WithTimeout(kafkaConsumerTimeout).WithPolling(projectorPoll).Should(Succeed())
		})
	}
})

var _ = Describe("KYC Kafka consumer — rejection", func() {
	ctx := context.Background()

	type kycRejectedInput struct {
		Currency string `yaml:"currency"`
		Reason   string `yaml:"reason"`
	}
	fixtures := helpers.LoadFixtures[kycRejectedInput](kycRejectedFS, "testdata/fixtures/kyc_rejected.yaml")

	for _, f := range fixtures {
		f := f
		It(f.Scenario, func() {
			accountID := domain.NewAccountID()
			customerID := domain.NewCustomerID()

			Expect(cmdBus.Dispatch(ctx, appaccount.OpenAccountCommand{
				AccountID:  accountID,
				CustomerID: customerID,
				Currency:   f.Input.Currency,
			})).To(Succeed())

			Eventually(func(g Gomega) {
				result, err := readRepo.GetBalance(ctx, accountID)
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(result.Status).To(Equal("Pending"))
			}).WithTimeout(projectorTimeout).WithPolling(projectorPoll).Should(Succeed())

			producer.PublishKYCRejected(ctx, GinkgoT(), string(customerID), f.Input.Reason)

			Eventually(func(g Gomega) {
				result, err := readRepo.GetBalance(ctx, accountID)
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(result.Status).To(Equal("Frozen"))
			}).WithTimeout(kafkaConsumerTimeout).WithPolling(projectorPoll).Should(Succeed())
		})
	}
})

// kafkaConsumerTimeout is longer than projectorTimeout to account for Kafka delivery latency.
const kafkaConsumerTimeout = 30 * helpers.Second

type kycConsumerInput struct {
	Currency string `yaml:"currency"`
}
