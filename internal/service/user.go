package service

import (
	"context"
	"errors"
	"strconv"

	"github.com/Quickaxe-Martina/gofermart/internal/storage"
)

// GetUserBalance get balance
func (s *Service) GetUserBalance(ctx context.Context, userID int) (*storage.UserBalance, error) {
	balance, err := s.store.GetBalanceByUser(ctx, userID)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &balance, nil
}

// WithdrawUser withdraw user
func (s *Service) WithdrawUser(ctx context.Context, userID int, sum float64, orderNumber string) error {
	orderCode, err := strconv.Atoi(orderNumber)
	if err != nil {
		return ErrIncorrectOrderNumber
	}

	if !LuhnValidate(orderCode) {
		return ErrIncorrectOrderNumber
	}

	err = s.store.WithdrawUser(ctx, userID, sum, orderNumber)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			return ErrUserNotFound
		} else if errors.Is(err, storage.ErrLowBalance) {
			return ErrLowBalance
		}
		return err
	}
	return nil
}

// GetUserWithdrawals get user withdrawals
func (s *Service) GetUserWithdrawals(ctx context.Context, userID int) ([]storage.Withdrawal, error) {
	withdrawals, err := s.store.GetWithdrawalsByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	return withdrawals, nil
}

// GetUserOrders get user orders
func (s *Service) GetUserOrders(ctx context.Context, userID int) ([]storage.Order, error) {
	orders, err := s.store.GetOrdersByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	return orders, nil
}

// CreateOrder create user order
func (s *Service) CreateOrder(ctx context.Context, orderNumber string, userID int) error {
	orderNumberInt, err := strconv.Atoi(orderNumber)
	if err != nil {
		return ErrIncorrectOrderNumber
	}

	if !LuhnValidate(orderNumberInt) {
		return ErrIncorrectOrderNumber
	}

	_, err = s.store.CreateOrder(ctx, orderNumber, userID)

	if err != nil {
		if errors.Is(err, storage.ErrOrderCreatedByAnotherUser) {
			return ErrOrderCreatedByAnotherUser
		} else if errors.Is(err, storage.ErrOrderAlreadyCreatedByUser) {
			return ErrOrderAlreadyCreatedByUser
		}
		return err
	}

	s.orderWorker.AddTask(orderNumber, userID)
	return nil
}

// CreateUser create new user
func (s *Service) CreateUser(ctx context.Context, username string, passwordHash string) (*storage.User, error) {
	user, err := s.store.CreateUser(ctx, username, passwordHash)
	if err != nil {
		if errors.Is(err, storage.ErrUserNameAlreadyExists) {
			return nil, ErrUserNameAlreadyExists
		}
		return nil, err
	}
	return &user, nil
}

// GetUser get user
func (s *Service) GetUser(ctx context.Context, username string) (*storage.User, error) {
	user, err := s.store.GetUserByUserName(ctx, username)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}
