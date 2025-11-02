package client

import (
	"context"
	"errors"
)

// ErrOrderNotRegistered order not registered
var ErrOrderNotRegistered = errors.New("order not registered")

// ErrTooManyRequest to many request
var ErrTooManyRequest = errors.New("to many request")

// AccrualResponse response model
type AccrualResponse struct {
	Order   string  `json:"order" validate:"required"`
	Status  string  `json:"status" validate:"required,oneof=REGISTERED INVALID PROCESSING PROCESSED"`
	Accrual float64 `json:"accrual"`
}

// AccrualClient interface for accrual service
type AccrualClient interface {
	GetOrder(context.Context, string) (AccrualResponse, error)
	Close() error
}
