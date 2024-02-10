package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi"
	"go.uber.org/zap"

	"github.com/AlexTerra21/gophermart/internal/app/auth"
	"github.com/AlexTerra21/gophermart/internal/app/config"
	"github.com/AlexTerra21/gophermart/internal/app/errs"
	"github.com/AlexTerra21/gophermart/internal/app/logger"
	"github.com/AlexTerra21/gophermart/internal/app/storage"
)

func MainRouter(c *config.Config) chi.Router {
	r := chi.NewRouter()
	r.Post("/api/user/register", logger.WithLogging(register(c)))
	r.Post("/api/user/login", logger.WithLogging(login(c)))
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

func register(c *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var user storage.User
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&user); err != nil {
			logger.Log().Debug("cannot decode request JSON body", zap.Error(err))
			http.Error(w, "", http.StatusInternalServerError)
			return
		}
		userID, err := c.Storage.AddUser(&user)
		if err != nil {
			logger.Log().Debug("Error adding new user", zap.Error(err))
			if errors.Is(err, errs.ErrConflict) {
				http.Error(w, "Login busy", http.StatusConflict)
				return
			}
			http.Error(w, "Error add user", http.StatusInternalServerError)
			return
		}
		cookie := &http.Cookie{
			Name: "Authorization",
		}
		token, err := auth.BuildJWTString(userID)
		if err != nil {
			logger.Log().Debug("Error generate token", zap.Error(err))
			http.Error(w, "Error generate token", http.StatusInternalServerError)
			return
		}
		cookie.Value = token
		http.SetCookie(w, cookie)

	}
}

func login(c *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var user storage.User
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&user); err != nil {
			logger.Log().Debug("cannot decode request JSON body", zap.Error(err))
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}
		userID, err := c.Storage.Authenticate(&user)
		if err != nil {
			logger.Log().Debug("Database error", zap.Error(err))
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
		if userID < 0 {
			http.Error(w, "Authenticate error", http.StatusUnauthorized)
			return
		}
		cookie := &http.Cookie{
			Name: "Authorization",
		}
		token, err := auth.BuildJWTString(userID)
		if err != nil {
			logger.Log().Debug("Error generate token", zap.Error(err))
			http.Error(w, "Error generate token", http.StatusInternalServerError)
			return
		}
		cookie.Value = token
		http.SetCookie(w, cookie)
	}

}
