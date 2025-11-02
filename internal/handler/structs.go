package handler

import (
	"github.com/Quickaxe-Martina/gofermart/internal/config"
	"github.com/Quickaxe-Martina/gofermart/internal/repository"
	"github.com/Quickaxe-Martina/gofermart/internal/storage"
	"go.uber.org/zap"
)

// Handler data
type Handler struct {
	cfg         *config.Config
	store       storage.Storage
	orderWorker *repository.OrderWorkers
	logging     *zap.Logger
}

// NewHandler create Handler
func NewHandler(cfg *config.Config, store storage.Storage, orderWorker *repository.OrderWorkers, logging *zap.Logger) *Handler {
	return &Handler{
		cfg:         cfg,
		store:       store,
		orderWorker: orderWorker,
		logging:     logging,
	}
}
