package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/savvinovan/kyc-service/internal/interfaces/http/handler"
)

func NewRouter(health *handler.HealthHandler, kyc *handler.KYCHandler) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)

	r.Get("/health", health.Handle)

	r.Route("/kyc", func(r chi.Router) {
		r.Post("/", kyc.Submit)
		r.Post("/{id}/approve", kyc.Approve)
		r.Post("/{id}/reject", kyc.Reject)
		r.Get("/{id}", kyc.GetStatus)
	})

	return r
}

var _ http.Handler = (*chi.Mux)(nil)
