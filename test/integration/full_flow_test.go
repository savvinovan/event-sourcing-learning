//go:build integration

// Package integration_test contains cross-service integration tests that exercise
// the full event-driven flow: wallet account → KYC submit → KYC approve → account
// activated via Kafka → withdrawal.
package integration_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/shopspring/decimal"

	walletappaccount "github.com/savvinovan/wallet-service/internal/application/account"
	walletdomain "github.com/savvinovan/wallet-service/internal/domain/account"

	kycapplication "github.com/savvinovan/kyc-service/internal/application/kyc"
	kycdomain "github.com/savvinovan/kyc-service/internal/domain/kyc"

	"github.com/savvinovan/event-sourcing-learning/tests/integration/helpers"
)

var _ = Describe("Full cross-service flow", func() {
	ctx := context.Background()

	It("open account → submit KYC → approve → account Active → withdraw", func() {
		accountID := walletdomain.NewAccountID()
		customerID := walletdomain.NewCustomerID()
		kycID := kycdomain.NewVerificationID()

		// 1. Open wallet account (status: Pending).
		Expect(walletCmdBus.Dispatch(ctx, walletappaccount.OpenAccountCommand{
			AccountID:  accountID,
			CustomerID: customerID,
			Currency:   "USD",
		})).To(Succeed())

		// Wait for projector to reflect Pending state.
		Eventually(func(g Gomega) {
			result, err := walletReadRepo.GetBalance(ctx, accountID)
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(result.Status).To(Equal("Pending"))
		}).WithTimeout(10 * helpers.Second).WithPolling(200 * helpers.Millisecond).Should(Succeed())

		// 2. Submit KYC verification for the same customer.
		Expect(kycCmdBus.Dispatch(ctx, kycapplication.SubmitKYCCommand{
			VerificationID: kycID,
			CustomerID:     kycdomain.CustomerID(customerID),
		})).To(Succeed())

		// 3. Approve KYC — publishes KYCVerified to Kafka.
		Expect(kycCmdBus.Dispatch(ctx, kycapplication.ApproveKYCCommand{
			VerificationID: kycID,
		})).To(Succeed())

		// 4. KYCVerified Kafka message → wallet KYC consumer → ActivateAccount command
		//    → AccountActivated event → projector updates read model to Active.
		Eventually(func(g Gomega) {
			result, err := walletReadRepo.GetBalance(ctx, accountID)
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(result.Status).To(Equal("Active"))
		}).WithTimeout(30 * helpers.Second).WithPolling(200 * helpers.Millisecond).Should(Succeed())

		// 5. Deposit money so we can withdraw.
		depositAmt, _ := decimal.NewFromString("500.00")
		depositMoney, _ := walletdomain.NewMoney(depositAmt, "USD")
		Expect(walletCmdBus.Dispatch(ctx, walletappaccount.DepositMoneyCommand{
			AccountID: accountID,
			Amount:    depositMoney,
		})).To(Succeed())

		Eventually(func(g Gomega) {
			result, err := walletReadRepo.GetBalance(ctx, accountID)
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(result.Balance.Amount.StringFixed(2)).To(Equal("500.00"))
		}).WithTimeout(10 * helpers.Second).WithPolling(200 * helpers.Millisecond).Should(Succeed())

		// 6. Withdraw money — succeeds because account is Active.
		withdrawAmt, _ := decimal.NewFromString("200.00")
		withdrawMoney, _ := walletdomain.NewMoney(withdrawAmt, "USD")
		Expect(walletCmdBus.Dispatch(ctx, walletappaccount.WithdrawMoneyCommand{
			AccountID: accountID,
			Amount:    withdrawMoney,
		})).To(Succeed())

		Eventually(func(g Gomega) {
			result, err := walletReadRepo.GetBalance(ctx, accountID)
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(result.Balance.Amount.StringFixed(2)).To(Equal("300.00"))
		}).WithTimeout(10 * helpers.Second).WithPolling(200 * helpers.Millisecond).Should(Succeed())
	})

	It("open account → submit KYC → reject → account Frozen", func() {
		accountID := walletdomain.NewAccountID()
		customerID := walletdomain.NewCustomerID()
		kycID := kycdomain.NewVerificationID()

		Expect(walletCmdBus.Dispatch(ctx, walletappaccount.OpenAccountCommand{
			AccountID:  accountID,
			CustomerID: customerID,
			Currency:   "USD",
		})).To(Succeed())

		Eventually(func(g Gomega) {
			result, err := walletReadRepo.GetBalance(ctx, accountID)
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(result.Status).To(Equal("Pending"))
		}).WithTimeout(10 * helpers.Second).WithPolling(200 * helpers.Millisecond).Should(Succeed())

		Expect(kycCmdBus.Dispatch(ctx, kycapplication.SubmitKYCCommand{
			VerificationID: kycID,
			CustomerID:     kycdomain.CustomerID(customerID),
		})).To(Succeed())

		Expect(kycCmdBus.Dispatch(ctx, kycapplication.RejectKYCCommand{
			VerificationID: kycID,
			Reason:         "documents expired",
		})).To(Succeed())

		Eventually(func(g Gomega) {
			result, err := walletReadRepo.GetBalance(ctx, accountID)
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(result.Status).To(Equal("Frozen"))
		}).WithTimeout(30 * helpers.Second).WithPolling(200 * helpers.Millisecond).Should(Succeed())
	})
})
