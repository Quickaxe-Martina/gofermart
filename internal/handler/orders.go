package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/Quickaxe-Martina/gofermart/internal/service"
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

	orderCode := string(body)

	err = h.service.CreateOrder(r.Context(), orderCode, user.ID)

	if err != nil {
		if errors.Is(err, service.ErrOrderCreatedByAnotherUser) {
			http.Error(w, "The order number has already been uploaded by another user", http.StatusConflict)
		} else if errors.Is(err, service.ErrOrderAlreadyCreatedByUser) {
			w.WriteHeader(http.StatusOK)
		} else if errors.Is(err, service.ErrIncorrectOrderNumber) {
			http.Error(w, "Incorrect order number", http.StatusUnprocessableEntity)
		} else {
			h.logging.Error("", zap.Error(err))
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

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
	orders, err := h.service.GetUserOrders(r.Context(), user.ID)
	if err != nil {
		h.logging.Error("", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}

	for _, order := range orders {
		responses = append(responses, UserOrdersResponse{
			Number:     order.OrderNumber,
			Status:     order.Status,
			Accrual:    order.Accrual,
			UploadedAt: order.UploadedAt,
		})
	}

	w.Header().Set("Content-Type", "application/json")

	enc := json.NewEncoder(w)
	if err := enc.Encode(responses); err != nil {
		h.logging.Error("error encoding response", zap.Error(err))
		return
	}

}
