package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/savvinovan/wallet-service/internal/interfaces/http/handler"
)

func NewRouter(health *handler.HealthHandler, account *handler.AccountHandler) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)

	r.Get("/health", health.Handle)

	r.Route("/accounts", func(r chi.Router) {
		r.Post("/", account.OpenAccount)
		r.Post("/{id}/deposit", account.Deposit)
		r.Post("/{id}/withdraw", account.Withdraw)
		r.Get("/{id}/balance", account.GetBalance)
		r.Get("/{id}/transactions", account.GetTransactions)
	})

	return r
}

// compile-time check
var _ http.Handler = (*chi.Mux)(nil)
