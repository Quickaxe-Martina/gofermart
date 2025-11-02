package storage

import (
	"github.com/Quickaxe-Martina/gofermart/internal/config"
	"go.uber.org/zap"
)

// NewStorage выбирает и возвращает реализацию интерфейса Storage
func NewStorage(cfg *config.Config, logging *zap.Logger) (Storage, error) {
	return NewPostgresStorage(cfg, logging)
}
