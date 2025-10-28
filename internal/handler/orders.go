package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/Quickaxe-Martina/gofermart/internal/logger"
	"github.com/Quickaxe-Martina/gofermart/internal/service"
	"github.com/Quickaxe-Martina/gofermart/internal/storage"
	"go.uber.org/zap"
)

// CreateOrder create order
func (h *Handler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r.Context())

	body, err := io.ReadAll(r.Body)
	if err != nil || len(body) == 0 {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	orderCode, err := strconv.Atoi(string(body))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if !service.LuhnValidate(orderCode) {
		http.Error(w, "Incorrect order number", http.StatusUnprocessableEntity)
		return
	}

	_, err = h.store.CreateOrder(r.Context(), orderCode, user.ID)

	if err != nil {
		if errors.Is(err, storage.ErrOrderCreatedByAnotherUser) {
			http.Error(w, "The order number has already been uploaded by another user", http.StatusConflict)
		} else if errors.Is(err, storage.ErrOrderAlreadyCreatedByUser) {
			w.WriteHeader(http.StatusOK)
			logger.Log.Error("", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	// TODO: add task

	w.WriteHeader(http.StatusAccepted)
}

// UserOrdersResponse model for response
type UserOrdersResponse struct {
	Number     string    `json:"number"`
	Status     string    `json:"status"`
	Accrual    float64   `json:"accrual"`
	UploadedAt time.Time `json:"uploaded_at"`
}

// GetOrders get orders
func (h *Handler) GetOrders(w http.ResponseWriter, r *http.Request) {
	var responses []UserOrdersResponse
	user := GetUser(r.Context())
	orders, err := h.store.GetOrdersByUser(r.Context(), user.ID)
	if err != nil {
		logger.Log.Error("", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}

	for _, order := range orders {
		responses = append(responses, UserOrdersResponse{
			Number:     strconv.Itoa(order.OrderNumber),
			Status:     order.Status,
			Accrual:    order.Accrual,
			UploadedAt: order.UploadedAt,
		})
	}

	w.Header().Set("Content-Type", "application/json")

	enc := json.NewEncoder(w)
	if err := enc.Encode(responses); err != nil {
		logger.Log.Error("error encoding response", zap.Error(err))
		return
	}

}
