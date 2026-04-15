//go:build integration

package integration_test

import (
	"context"
	"embed"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/shopspring/decimal"

	appaccount "github.com/savvinovan/wallet-service/internal/application/account"
	domain "github.com/savvinovan/wallet-service/internal/domain/account"
	"github.com/savvinovan/wallet-service/test/integration/helpers"
)

//go:embed testdata/fixtures/open_account.yaml
var openAccountFS embed.FS

//go:embed testdata/fixtures/deposit_withdraw.yaml
var depositWithdrawFS embed.FS

type openAccountInput struct {
	Currency string `yaml:"currency"`
}

type depositWithdrawInput struct {
	Currency       string `yaml:"currency"`
	DepositAmount  string `yaml:"deposit_amount"`
	WithdrawAmount string `yaml:"withdraw_amount,omitempty"`
}

var _ = Describe("Account", func() {
	ctx := context.Background()

	Describe("Opening an account", func() {
		fixtures := helpers.LoadFixtures[openAccountInput](openAccountFS, "testdata/fixtures/open_account.yaml")

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

				exp := mapToStruct(f.Expected)
				Eventually(func(g Gomega) {
					result, err := readRepo.GetBalance(ctx, accountID)
					g.Expect(err).NotTo(HaveOccurred())
					g.Expect(result.Status).To(Equal(exp["status"]))
					g.Expect(result.Balance.Amount.StringFixed(2)).To(Equal(exp["balance"]))
				}).WithTimeout(projectorTimeout).WithPolling(projectorPoll).Should(Succeed())
			})
		}
	})

	Describe("Deposit and withdraw", func() {
		fixtures := helpers.LoadFixtures[depositWithdrawInput](depositWithdrawFS, "testdata/fixtures/deposit_withdraw.yaml")

		for _, f := range fixtures {
			f := f
			It(f.Scenario, func() {
				accountID := domain.NewAccountID()
				customerID := domain.NewCustomerID()
				exp := mapToStruct(f.Expected)

				// Open and activate so deposit/withdraw are allowed.
				Expect(cmdBus.Dispatch(ctx, appaccount.OpenAccountCommand{
					AccountID: accountID, CustomerID: customerID, Currency: f.Input.Currency,
				})).To(Succeed())
				Expect(cmdBus.Dispatch(ctx, appaccount.ActivateAccountCommand{AccountID: accountID})).To(Succeed())

				depositAmt, _ := decimal.NewFromString(f.Input.DepositAmount)
				depositMoney, _ := domain.NewMoney(depositAmt, f.Input.Currency)
				Expect(cmdBus.Dispatch(ctx, appaccount.DepositMoneyCommand{
					AccountID: accountID, Amount: depositMoney,
				})).To(Succeed())

				Eventually(func(g Gomega) {
					result, err := readRepo.GetBalance(ctx, accountID)
					g.Expect(err).NotTo(HaveOccurred())
					g.Expect(result.Balance.Amount.StringFixed(2)).To(Equal(exp["balance_after_deposit"]))
				}).WithTimeout(projectorTimeout).WithPolling(projectorPoll).Should(Succeed())

				if f.Input.WithdrawAmount == "" {
					return
				}

				withdrawAmt, _ := decimal.NewFromString(f.Input.WithdrawAmount)
				withdrawMoney, _ := domain.NewMoney(withdrawAmt, f.Input.Currency)
				err := cmdBus.Dispatch(ctx, appaccount.WithdrawMoneyCommand{
					AccountID: accountID, Amount: withdrawMoney,
				})

				if exp["error"] == "insufficient_funds" {
					Expect(err).To(MatchError(domain.ErrInsufficientFunds))
					return
				}
				Expect(err).To(Succeed())

				Eventually(func(g Gomega) {
					result, err := readRepo.GetBalance(ctx, accountID)
					g.Expect(err).NotTo(HaveOccurred())
					g.Expect(result.Balance.Amount.StringFixed(2)).To(Equal(exp["balance_after_withdraw"]))
				}).WithTimeout(projectorTimeout).WithPolling(projectorPoll).Should(Succeed())
			})
		}
	})
})

// projectorTimeout / projectorPoll bound all Eventually assertions that wait for
// the async projector to update the read model.
const (
	projectorTimeout = 10 * helpers.Second
	projectorPoll    = 200 * helpers.Millisecond
)

// mapToStruct converts the untyped map that yaml.v3 produces for the Expected field
// into map[string]string for easy access in specs.
func mapToStruct(raw any) map[string]string {
	m, _ := raw.(map[string]any)
	result := make(map[string]string, len(m))
	for k, v := range m {
		if v == nil {
			continue
		}
		result[k] = v.(string)
	}
	return result
}
