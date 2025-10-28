package client

import (
	"context"
	"errors"
)

// ErrOrderNotRegistered todo
var ErrOrderNotRegistered = errors.New("order not registered")

// ErrToManyRequests todo
var ErrToManyRequests = errors.New("to many requests")

// AccrualResponse todo
type AccrualResponse struct {
	Order   string  `json:"order" validate:"required"`
	Status  string  `json:"status" validate:"required,oneof=REGISTERED INVALID PROCESSING PROCESSED"`
	Accrual float64 `json:"accrual"`
}

// AccrualClient todo
type AccrualClient interface {
	GetOrder(context.Context, string) (AccrualResponse, error)
	Close() error
}
