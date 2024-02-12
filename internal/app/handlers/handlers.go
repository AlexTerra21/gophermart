package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi"
	"go.uber.org/zap"

	"github.com/AlexTerra21/gophermart/internal/app/auth"
	"github.com/AlexTerra21/gophermart/internal/app/compress"
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
	r.Get("/api/user/orders", logger.WithLogging(getOrders(c)))
	r.Get("/api/user/orders", compress.WithCompress(logger.WithLogging(getOrders(c))))
	r.Get("/api/user/balance", compress.WithCompress(logger.WithLogging(balance(c))))
	r.Post("/api/user/balance/withdraw", logger.WithLogging(withdraw(c)))
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
		userID, err := c.Storage.AddUser(r.Context(), &user)
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
		userID, err := c.Storage.CheckLoginPassword(r.Context(), &user)
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
		number, err := strconv.ParseInt(string(strNumber), 10, 64)
		if err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest) // 400
			return
		}
		if !utils.Valid(number) {
			http.Error(w, "Invalid number", http.StatusUnprocessableEntity) // 422
			return
		}

		order, err := c.Storage.SetOrder(r.Context(), number, userID)
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
		c.OrderQueue.Push(order)
		w.WriteHeader(http.StatusAccepted) // 202
	}
}

func getOrders(c *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := auth.CheckAuth(r)
		if userID < 0 {
			w.WriteHeader(http.StatusUnauthorized) // 401
			return
		}
		orders, err := c.Storage.GetOrders(r.Context(), userID)
		if err != nil {
			logger.Log().Debug("Error getting orders", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError) // 500
			return
		}
		if len(orders) == 0 {
			w.WriteHeader(http.StatusNoContent) // 204
			return
		}
		w.Header().Set("content-type", "application/json")
		w.WriteHeader(http.StatusOK) //200
		encoder := json.NewEncoder(w)
		if err := encoder.Encode(orders); err != nil {
			logger.Log().Debug("error encoding response", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError) // 500
			return
		}
	}
}

func balance(c *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := auth.CheckAuth(r)
		if userID < 0 {
			w.WriteHeader(http.StatusUnauthorized) // 401
			return
		}

		sumAccrual, err := c.Storage.GetBalance(r.Context(), userID)
		if err != nil {
			logger.Log().Debug("Error getting balance", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError) // 500
			return
		}

		sumWithdraw, err := c.Storage.GetWithdrawSum(r.Context(), userID)
		if err != nil {
			logger.Log().Debug("Error getting withdraw sum", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError) // 500
			return
		}

		withdraw := storage.Withdrawal{
			Current:   sumAccrual - sumWithdraw,
			Withdrawn: sumWithdraw,
		}

		w.Header().Set("content-type", "application/json")
		w.WriteHeader(http.StatusOK) // 200
		encoder := json.NewEncoder(w)
		if err := encoder.Encode(&withdraw); err != nil {
			logger.Log().Debug("error encoding response", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError) // 500
			return
		}
	}
}

func withdraw(c *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := auth.CheckAuth(r)
		if userID < 0 {
			w.WriteHeader(http.StatusUnauthorized) // 401
			return
		}
		withdrawRequest := storage.WithdrawRequest{}
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&withdrawRequest); err != nil {
			logger.Log().Debug("cannot decode request JSON body", zap.Error(err))
			http.Error(w, "Bad request", http.StatusBadRequest) // 400
			return
		}

		number, err := strconv.ParseInt(string(withdrawRequest.Order), 10, 64)
		if err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest) // 400
			return
		}
		if !utils.Valid(number) {
			http.Error(w, "Invalid number", http.StatusUnprocessableEntity) // 422
			return
		}

		sumAccrual, err := c.Storage.GetBalance(r.Context(), userID)
		if err != nil {
			logger.Log().Debug("Error getting balance", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError) // 500
			return
		}

		if sumAccrual < withdrawRequest.Sum {
			http.Error(w, "Payment Required", http.StatusPaymentRequired) // 402
			return
		}
		withdraw := storage.Withdrawal{
			UserID:      userID,
			Order:       withdrawRequest.Order,
			Withdrawn:   withdrawRequest.Sum,
			ProcessedAt: time.Now(),
		}
		err = c.Storage.SetWithdraw(r.Context(), withdraw)
		if err != nil {
			logger.Log().Debug("Error add withdraw", zap.Error(err))
			if errors.Is(err, errs.ErrConflict) {
				http.Error(w, "Conflict", http.StatusConflict) //409
				return
			} else {
				w.WriteHeader(http.StatusInternalServerError) // 500
				return
			}
		}

	}
}
