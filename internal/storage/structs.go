package storage

import (
	"context"
	"errors"
)

// ErrUserNameAlreadyExists username is already taken
var ErrUserNameAlreadyExists = errors.New("username is already taken")

// ErrUserNotFound user not found
var ErrUserNotFound = errors.New("user not found")

// ErrNotImplemented not implemented
var ErrNotImplemented = errors.New("not implemented")

// User model
type User struct {
	ID           int
	UserName     string
	PasswordHash string
}

// UserStorage defines methods for user management
type UserStorage interface {
	CreateUser(ctx context.Context, username string, passwordHash string) (User, error)
	GetUserByUserName(ctx context.Context, username string) (User, error)
}

// Storage defines methods
type Storage interface {
	UserStorage
	Close() error
	Ping(ctx context.Context) error
}
