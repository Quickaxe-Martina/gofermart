package storage

import (
	"context"
	"errors"
	"time"
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

// ErrOrderCreatedByAnotherUser todo
var ErrOrderCreatedByAnotherUser = errors.New("order cteated by another user")

// ErrOrderAlreadyCreatedByUser todo
var ErrOrderAlreadyCreatedByUser = errors.New("order is already created")

const (
	StatusNew        = "NEW"
	StatusProcessing = "PROCESSING"
	StatusInvalid    = "INVALID"
	StatusProcessed  = "PROCESSED"
)

// Order model
type Order struct {
	ID          int
	OrderNumber int
	Status      string
	Accrual     float64
	UserID      int
	UploadedAt  time.Time
}

// OrderStorage defines methods for order management
type OrderStorage interface {
	CreateOrder(ctx context.Context, orderNumber int, userID int) (Order, error)
	GetOrdersByUser(ctx context.Context, userID int) ([]Order, error)
}

// User model
type User struct {
	ID           int
	UserName     string
	PasswordHash string
}

// UserBalance todo
type UserBalance struct {
	Balance   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

// Withdrawal todo
type Withdrawal struct {
	OrderNumber int64
	Sum         float64
	CreatedAt   time.Time
}

// UserStorage defines methods for user management
type UserStorage interface {
	CreateUser(ctx context.Context, username string, passwordHash string) (User, error)
	GetUserByUserName(ctx context.Context, username string) (User, error)
	GetBalanceByUser(ctx context.Context, userID int) (UserBalance, error)
	WithdrawUser(ctx context.Context, userID int, sum float64, orderNumber int) error
	GetWithdrawalsByUser(ctx context.Context, userID int) ([]Withdrawal, error)
	AccrueUser(ctx context.Context, userID int, sum float64, status string, orderNumber string) error
}

// Storage defines methods
type Storage interface {
	UserStorage
	OrderStorage
	Close() error
	Ping(ctx context.Context) error
}
