package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"go.uber.org/zap"

	"github.com/AlexTerra21/gophermart/internal/app/auth"
	"github.com/AlexTerra21/gophermart/internal/app/config"
	"github.com/AlexTerra21/gophermart/internal/app/errs"
	"github.com/AlexTerra21/gophermart/internal/app/logger"
	"github.com/AlexTerra21/gophermart/internal/app/storage"
	"github.com/AlexTerra21/gophermart/internal/app/utils"
)

func MainRouter(c *config.Config) chi.Router {
	r := chi.NewRouter()
	r.Post("/api/user/register", logger.WithLogging(register(c)))
	r.Post("/api/user/login", logger.WithLogging(login(c)))
	r.Post("/api/user/orders", logger.WithLogging(addOrder(c)))
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
			http.Error(w, "", http.StatusInternalServerError) // 500
			return
		}
		userID, err := c.Storage.AddUser(&user)
		if err != nil {
			logger.Log().Debug("Error adding new user", zap.Error(err))
			if errors.Is(err, errs.ErrConflict) {
				http.Error(w, "Login busy", http.StatusConflict) // 409
				return
			}
			http.Error(w, "Error add user", http.StatusInternalServerError) // 500
			return
		}
		cookie := &http.Cookie{
			Name: "Authorization",
		}
		token, err := auth.BuildJWTString(userID)
		if err != nil {
			logger.Log().Debug("Error generate token", zap.Error(err))
			http.Error(w, "Error generate token", http.StatusInternalServerError) // 500
			return
		}
		cookie.Value = token
		http.SetCookie(w, cookie) // 200
	}
}

func login(c *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var user storage.User
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&user); err != nil {
			logger.Log().Debug("cannot decode request JSON body", zap.Error(err))
			http.Error(w, "Bad request", http.StatusBadRequest) // 400
			return
		}
		userID, err := c.Storage.CheckLoginPassword(&user)
		if err != nil {
			logger.Log().Debug("Database error", zap.Error(err))
			http.Error(w, "Database error", http.StatusInternalServerError) // 500
			return
		}
		if userID < 0 {
			http.Error(w, "Invalid login or password", http.StatusUnauthorized) // 401
			return
		}
		cookie := &http.Cookie{
			Name: "Authorization",
		}
		token, err := auth.BuildJWTString(userID)
		if err != nil {
			logger.Log().Debug("Error generate token", zap.Error(err))
			http.Error(w, "Error generate token", http.StatusInternalServerError) // 500
			return
		}
		cookie.Value = token
		http.SetCookie(w, cookie) // 200
	}

}

func addOrder(c *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := auth.CheckAuth(r)
		if userID < 0 {
			http.Error(w, "Unauthorized", http.StatusUnauthorized) // 401
			return
		}
		strNumber, _ := io.ReadAll(r.Body)
		number, err := strconv.Atoi(string(strNumber))
		if err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest) // 400
			return
		}
		if !utils.Valid(number) {
			http.Error(w, "Invalid number", http.StatusUnprocessableEntity) // 422
			return
		}
		order, err := c.Storage.SetOrder(number, userID)
		if err != nil {
			logger.Log().Debug("Error adding new order", zap.Error(err))
			if errors.Is(err, errs.ErrConflict) {
				if userID == order.UserID {
					w.WriteHeader(http.StatusOK) // 200
					return
				} else {
					http.Error(w, "Conflict", http.StatusConflict) //409
					return
				}
			} else {
				http.Error(w, "Error adding new order", http.StatusInternalServerError) // 500
				return
			}
		}
		w.WriteHeader(http.StatusAccepted) // 202
	}
}
