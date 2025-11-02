package handler

import (
	"context"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// Ping db
func (h *Handler) Ping(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	if err := h.store.Ping(ctx); err != nil {
		h.logging.Error("db error", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.WriteHeader(http.StatusOK)
}
