// Package app assembles the wallet-service API application via uber/fx.
package app

import (
	"go.uber.org/fx"

	appaccount "github.com/savvinovan/wallet-service/internal/application/account"
	"github.com/savvinovan/wallet-service/internal/application/command"
	"github.com/savvinovan/wallet-service/internal/application/query"
	"github.com/savvinovan/wallet-service/internal/interfaces/http/handler"
	httpinterface "github.com/savvinovan/wallet-service/internal/interfaces/http"
)

// New constructs the fx application graph for the wallet API.
func New() *fx.App {
	return fx.New(
		fx.Provide(
			// Config & logger
			newConfig,
			newLogger,

			// Infrastructure
			newDBPool,
			newEventStore,
			newReadModelRepo,

			// Command handlers
			appaccount.NewOpenAccountHandler,
			appaccount.NewDepositMoneyHandler,
			appaccount.NewWithdrawMoneyHandler,
			appaccount.NewActivateAccountHandler,
			appaccount.NewFreezeAccountHandler,

			// Query handlers
			appaccount.NewGetBalanceHandler,
			appaccount.NewGetTransactionsHandler,

			// Buses
			newCommandBus,
			newQueryBus,

			// HTTP
			handler.NewHealthHandler,
			handler.NewAccountHandler,
			httpinterface.NewRouter,
			newHTTPServer,
		),
		fx.Invoke(runMigrations),
		fx.Invoke(startHTTPServer),
	)
}

func newCommandBus(
	openAccount *appaccount.OpenAccountHandler,
	deposit *appaccount.DepositMoneyHandler,
	withdraw *appaccount.WithdrawMoneyHandler,
	activate *appaccount.ActivateAccountHandler,
	freeze *appaccount.FreezeAccountHandler,
) command.Bus {
	bus := command.NewInMemoryBus()
	command.MustRegister(bus, openAccount)
	command.MustRegister(bus, deposit)
	command.MustRegister(bus, withdraw)
	command.MustRegister(bus, activate)
	command.MustRegister(bus, freeze)
	return bus
}

func newQueryBus(
	getBalance *appaccount.GetBalanceHandler,
	getTransactions *appaccount.GetTransactionsHandler,
) query.Bus {
	bus := query.NewInMemoryBus()
	query.MustRegister(bus, getBalance)
	query.MustRegister(bus, getTransactions)
	return bus
}
