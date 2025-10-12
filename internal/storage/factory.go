package storage

import (
	"github.com/Quickaxe-Martina/gofermart/internal/config"
)

// NewStorage выбирает и возвращает реализацию интерфейса Storage
func NewStorage(cfg *config.Config) (Storage, error) {
	return NewPostgresStorage(cfg), nil
}
