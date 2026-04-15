//go:build integration

package integration_test

import (
	"context"
	"embed"
	"errors"
	"sync"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/shopspring/decimal"

	appaccount "github.com/savvinovan/wallet-service/internal/application/account"
	domain "github.com/savvinovan/wallet-service/internal/domain/account"
	"github.com/savvinovan/wallet-service/internal/infrastructure/eventstore"
	"github.com/savvinovan/wallet-service/test/integration/helpers"
)

//go:embed testdata/fixtures/concurrent_withdrawals.yaml
var concurrentWithdrawalsFS embed.FS

var _ = Describe("Concurrent withdrawals", func() {
	ctx := context.Background()
	fixtures := helpers.LoadFixtures[concurrentWithdrawInput](concurrentWithdrawalsFS, "testdata/fixtures/concurrent_withdrawals.yaml")

	for _, f := range fixtures {
		f := f
		It(f.Scenario, func() {
			accountID := domain.NewAccountID()
			customerID := domain.NewCustomerID()

			// Set up a funded, active account.
			Expect(cmdBus.Dispatch(ctx, appaccount.OpenAccountCommand{
				AccountID: accountID, CustomerID: customerID, Currency: f.Input.Currency,
			})).To(Succeed())
			Expect(cmdBus.Dispatch(ctx, appaccount.ActivateAccountCommand{AccountID: accountID})).To(Succeed())

			depositAmt, _ := decimal.NewFromString(f.Input.DepositAmount)
			depositMoney, _ := domain.NewMoney(depositAmt, f.Input.Currency)
			Expect(cmdBus.Dispatch(ctx, appaccount.DepositMoneyCommand{
				AccountID: accountID, Amount: depositMoney,
			})).To(Succeed())

			// Fire two concurrent withdrawals.
			withdrawAmt, _ := decimal.NewFromString(f.Input.WithdrawAmount)
			withdrawMoney, _ := domain.NewMoney(withdrawAmt, f.Input.Currency)

			var mu sync.Mutex
			errs := make([]error, 2)
			var wg sync.WaitGroup
			wg.Add(2)
			for i := 0; i < 2; i++ {
				i := i
				go func() {
					defer wg.Done()
					err := cmdBus.Dispatch(ctx, appaccount.WithdrawMoneyCommand{
						AccountID: accountID, Amount: withdrawMoney,
					})
					mu.Lock()
					errs[i] = err
					mu.Unlock()
				}()
			}
			wg.Wait()

			successes := countBy(errs, func(e error) bool { return e == nil })
			failures := countBy(errs, func(e error) bool {
				return errors.Is(e, eventstore.ErrVersionConflict) ||
					errors.Is(e, domain.ErrInsufficientFunds)
			})

			Expect(successes).To(Equal(1), "exactly one withdrawal must succeed")
			Expect(failures).To(Equal(1), "exactly one withdrawal must fail (ErrVersionConflict or ErrInsufficientFunds)")
		})
	}
})

type concurrentWithdrawInput struct {
	Currency       string `yaml:"currency"`
	DepositAmount  string `yaml:"deposit_amount"`
	WithdrawAmount string `yaml:"withdraw_amount"`
}

func countBy[T any](slice []T, pred func(T) bool) int {
	n := 0
	for _, v := range slice {
		if pred(v) {
			n++
		}
	}
	return n
}
