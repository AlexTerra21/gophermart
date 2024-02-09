package handlers

import (
	"net/http"

	"github.com/go-chi/chi"

	"github.com/AlexTerra21/gophermart/internal/app/config"
	"github.com/AlexTerra21/gophermart/internal/app/logger"
)

func MainRouter(c *config.Config) chi.Router {
	r := chi.NewRouter()
	r.Post("/api/user/register", logger.WithLogging(empty(c)))
	r.Post("/api/user/login", logger.WithLogging(empty(c)))
	r.Post("/api/user/orders", logger.WithLogging(empty(c)))
	r.Get("/api/user/orders", logger.WithLogging(empty(c)))
	r.Get("/api/user/balance", logger.WithLogging(empty(c)))
	r.Post("/api/user/balance/withdraw", logger.WithLogging(empty(c)))
	r.Get("/api/user/withdrawals", logger.WithLogging(empty(c)))
	r.MethodNotAllowed(notAllowedHandler)
	return r
}

func notAllowedHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Unsupported method", http.StatusMethodNotAllowed) // В ответе код 400
}

func empty(c *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
	}
}
