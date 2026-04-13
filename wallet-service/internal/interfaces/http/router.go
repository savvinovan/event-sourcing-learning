package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/savvinovan/wallet-service/internal/interfaces/http/handler"
)

func NewRouter(health *handler.HealthHandler) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)

	r.Get("/health", health.Handle)

	return r
}

// compile-time check
var _ http.Handler = (*chi.Mux)(nil)
