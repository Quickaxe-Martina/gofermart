package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/Quickaxe-Martina/gofermart/internal/logger"
	"go.uber.org/zap"
)

// Ping db
func (h *Handler) Ping(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	if err := h.store.Ping(ctx); err != nil {
		logger.Log.Error("db error", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.WriteHeader(http.StatusOK)
}
