package http

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/savvinovan/wallet-service/internal/interfaces/http/gen"
	"github.com/savvinovan/wallet-service/internal/interfaces/http/handler"
)

func NewRouter(health *handler.HealthHandler, account *handler.AccountHandler, log *slog.Logger) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)

	r.Get("/health", health.Handle)

	strict := gen.NewStrictHandlerWithOptions(account, nil, gen.StrictHTTPServerOptions{
		RequestErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"message":"` + err.Error() + `"}`))
		},
		ResponseErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
			log.Error("response error", "error", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"message":"internal error"}`))
		},
	})
	gen.HandlerFromMux(strict, r)

	return r
}

// compile-time check
var _ http.Handler = (*chi.Mux)(nil)
