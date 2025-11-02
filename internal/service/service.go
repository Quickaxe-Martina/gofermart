package service

import (
	"github.com/Quickaxe-Martina/gofermart/internal/config"
	"github.com/Quickaxe-Martina/gofermart/internal/repository"
	"github.com/Quickaxe-Martina/gofermart/internal/storage"

	"errors"

	"go.uber.org/zap"
)

// ErrUserNameAlreadyExists username is already taken
var ErrUserNameAlreadyExists = errors.New("username is already taken")

// ErrUserNotFound user not found
var ErrUserNotFound = errors.New("user not found")

// ErrLowBalance low balance
var ErrLowBalance = errors.New("low balance")

// ErrNotImplemented not implemented
var ErrNotImplemented = errors.New("not implemented")

// ErrOrderAlreadyExists order is already taken
var ErrOrderAlreadyExists = errors.New("order is already taken")

// ErrOrderCreatedByAnotherUser order cteated by another user
var ErrOrderCreatedByAnotherUser = errors.New("order cteated by another user")

// ErrOrderAlreadyCreatedByUser order is already created
var ErrOrderAlreadyCreatedByUser = errors.New("order is already created")

// ErrIncorrectOrderNumber incorrect order number
var ErrIncorrectOrderNumber = errors.New("incorrect order number")

// Service Service layer
type Service struct {
	cfg         *config.Config
	store       storage.Storage
	orderWorker *repository.OrderWorkers
	logging     *zap.Logger
}

// ServiceConfig config for NewService
type ServiceConfig struct {
	Cfg         *config.Config
	Store       storage.Storage
	OrderWorker *repository.OrderWorkers
	Logging     *zap.Logger
}

// NewService create Service
func NewService(serviceCfg *ServiceConfig) *Service {
	return &Service{
		cfg:         serviceCfg.Cfg,
		store:       serviceCfg.Store,
		orderWorker: serviceCfg.OrderWorker,
		logging:     serviceCfg.Logging,
	}
}
