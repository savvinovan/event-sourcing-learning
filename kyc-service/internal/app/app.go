// Package app assembles the kyc-service API application via uber/fx.
package app

import (
	"go.uber.org/fx"

	appkyc "github.com/savvinovan/kyc-service/internal/application/kyc"
	"github.com/savvinovan/kyc-service/internal/application/command"
	"github.com/savvinovan/kyc-service/internal/application/query"
	"github.com/savvinovan/kyc-service/internal/interfaces/http/handler"
	httpinterface "github.com/savvinovan/kyc-service/internal/interfaces/http"
)

// New constructs the fx application graph for the KYC API.
func New() *fx.App {
	return fx.New(
		fx.Provide(
			// Config & logger
			newConfig,
			newLogger,

			// Infrastructure
			newEventStore,
			newPublisher,

			// Command handlers
			appkyc.NewSubmitKYCHandler,
			appkyc.NewApproveKYCHandler,
			appkyc.NewRejectKYCHandler,

			// Query handlers
			appkyc.NewGetKYCStatusHandler,

			// Buses
			newCommandBus,
			newQueryBus,

			// HTTP
			handler.NewHealthHandler,
			handler.NewKYCHandler,
			httpinterface.NewRouter,
			newHTTPServer,
		),
		fx.Invoke(startHTTPServer),
		fx.Invoke(registerPublisherShutdown),
	)
}

func newCommandBus(
	submit *appkyc.SubmitKYCHandler,
	approve *appkyc.ApproveKYCHandler,
	reject *appkyc.RejectKYCHandler,
) command.Bus {
	bus := command.NewInMemoryBus()
	command.MustRegister(bus, submit)
	command.MustRegister(bus, approve)
	command.MustRegister(bus, reject)
	return bus
}

func newQueryBus(
	getStatus *appkyc.GetKYCStatusHandler,
) query.Bus {
	bus := query.NewInMemoryBus()
	query.MustRegister(bus, getStatus)
	return bus
}
