//go:build integration

package integration_test

import (
	"context"
	"embed"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	appkyc "github.com/savvinovan/kyc-service/internal/application/kyc"
	domain "github.com/savvinovan/kyc-service/internal/domain/kyc"
	"github.com/savvinovan/kyc-service/test/integration/helpers"
)

//go:embed testdata/fixtures/submit_kyc.yaml
var submitKYCFS embed.FS

type submitKYCInput struct {
	CustomerIDSuffix string `yaml:"customer_id_suffix"`
}

var _ = Describe("KYC flow", func() {
	ctx := context.Background()

	Describe("Submitting a KYC verification", func() {
		fixtures := helpers.LoadFixtures[submitKYCInput](submitKYCFS, "testdata/fixtures/submit_kyc.yaml")

		for _, f := range fixtures {
			f := f
			It(f.Scenario, func() {
				verificationID := domain.NewVerificationID()
				customerID := domain.CustomerID(f.Input.CustomerIDSuffix)

				Expect(cmdBus.Dispatch(ctx, appkyc.SubmitKYCCommand{
					VerificationID: verificationID,
					CustomerID:     customerID,
				})).To(Succeed())

				// Verify event is stored by querying via query bus.
				result, err := qryBus.Ask(ctx, appkyc.GetKYCStatusQuery{
					VerificationID: verificationID,
				})
				Expect(err).NotTo(HaveOccurred())
				status, ok := result.(appkyc.KYCStatusResult)
				Expect(ok).To(BeTrue())
				Expect(string(status.CustomerID)).To(Equal(f.Input.CustomerIDSuffix))
				Expect(status.Status).To(Equal("Submitted"))
			})
		}
	})
})
