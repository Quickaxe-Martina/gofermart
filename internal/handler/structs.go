package handler

import (
	"github.com/Quickaxe-Martina/gofermart/internal/config"
	"github.com/Quickaxe-Martina/gofermart/internal/storage"
)

// Handler data
type Handler struct {
	cfg   *config.Config
	store storage.Storage
}

// NewHandler create Handler
func NewHandler(cfg *config.Config, store storage.Storage) *Handler {
	return &Handler{
		cfg:   cfg,
		store: store,
	}
}
