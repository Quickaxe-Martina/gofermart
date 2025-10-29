package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"strconv"
	"time"

	"github.com/Quickaxe-Martina/gofermart/internal/logger"
	"github.com/Quickaxe-Martina/gofermart/internal/service"
	"github.com/Quickaxe-Martina/gofermart/internal/storage"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

// GetUserBalance get balance
func (h *Handler) GetUserBalance(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r.Context())
	w.Header().Set("Content-Type", "application/json")
	balance, err := h.store.GetBalanceByUser(r.Context(), user.ID)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			logger.Log.Error(user.UserName)
			http.Error(w, "Invalid username", http.StatusUnauthorized)
		} else {
			logger.Log.Error("", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	enc := json.NewEncoder(w)
	if err := enc.Encode(balance); err != nil {
		logger.Log.Error("error encoding response", zap.Error(err))
		return
	}
	w.WriteHeader(http.StatusOK)
}

type withdrawUserRequest struct {
	Order string  `json:"order" validate:"required,numeric"`
	Sum   float64 `json:"sum" validate:"required,gte=0"`
}

// WithdrawUser todo
func (h *Handler) WithdrawUser(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r.Context())

	var req withdrawUserRequest
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&req); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	validate := validator.New(validator.WithRequiredStructEnabled())
	if err := validate.Struct(req); err != nil {
		logger.Log.Info("validation error", zap.Error(err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	orderCode, err := strconv.Atoi(req.Order)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if !service.LuhnValidate(orderCode) {
		http.Error(w, "Incorrect order number", http.StatusUnprocessableEntity)
		return
	}
	err = h.store.WithdrawUser(r.Context(), user.ID, req.Sum, orderCode)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			logger.Log.Error(user.UserName)
			http.Error(w, "Invalid username", http.StatusUnauthorized)
		} else if errors.Is(err, storage.ErrLowBalance) {
			http.Error(w, "Payment required", http.StatusPaymentRequired)
		} else {
			logger.Log.Error("", zap.Error(err))
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
}

// GetWithdrawalsByUserResponse todo
type GetWithdrawalsByUserResponse struct {
	Order       string  `json:"order"`
	Sum         float64 `json:"sum"`
	ProcessedAt string  `json:"processed_at"`
}

// GetWithdrawalsByUser todo
func (h *Handler) GetWithdrawalsByUser(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r.Context())
	withdrawals, err := h.store.GetWithdrawalsByUser(r.Context(), user.ID)
	if err != nil {
		logger.Log.Error("", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "application/json")

	if len(withdrawals) == 0 {
		w.WriteHeader(http.StatusNoContent)
	} else {
		resp := make([]GetWithdrawalsByUserResponse, 0, len(withdrawals))
		for _, wd := range withdrawals {
			resp = append(resp, GetWithdrawalsByUserResponse{
				Order:       strconv.Itoa(wd.OrderNumber),
				Sum:         wd.Sum,
				ProcessedAt: wd.CreatedAt.Format(time.RFC3339),
			})
		}
		enc := json.NewEncoder(w)
		if err := enc.Encode(resp); err != nil {
			logger.Log.Error("error encoding response", zap.Error(err))
			return
		}
		w.WriteHeader(http.StatusOK)
	}

}
