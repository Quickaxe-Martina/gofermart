package handler

import (
	"github.com/Quickaxe-Martina/gofermart/internal/config"
	"github.com/Quickaxe-Martina/gofermart/internal/service"
	"go.uber.org/zap"
)

// Handler data
type Handler struct {
	cfg     *config.Config
	service service.Service
	logging *zap.Logger
}

// NewHandler create Handler
func NewHandler(cfg *config.Config, service service.Service, logging *zap.Logger) *Handler {
	return &Handler{
		cfg:     cfg,
		service: service,
		logging: logging,
	}
}
