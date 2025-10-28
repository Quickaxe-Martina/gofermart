package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib" // driver

	"github.com/Quickaxe-Martina/gofermart/internal/config"
	_ "github.com/Quickaxe-Martina/gofermart/internal/logger" // logger
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

// PostgresStorage is DB implementation of the Storage interface
type PostgresStorage struct {
	DB *sql.DB
}

// NewPostgresStorage creates new PostgresStorage
func NewPostgresStorage(cfg *config.Config) *PostgresStorage {
	db, err := sql.Open("pgx", cfg.DatabaseDsn)
	if err != nil {
		panic(err)
	}
	if err := runMigrations(db, cfg.MigrationsPath); err != nil {
		panic(fmt.Errorf("failed to run migrations: %w", err))
	}
	store := &PostgresStorage{
		DB: db,
	}
	return store
}

// Close releases resources
func (store *PostgresStorage) Close() error {
	store.DB.Close()
	return nil
}

// Ping DB
func (store *PostgresStorage) Ping(ctx context.Context) error {
	if err := store.DB.PingContext(ctx); err != nil {
		return err
	}
	return nil
}

// CreateUser creates a new user and returns it
func (store *PostgresStorage) CreateUser(ctx context.Context, username string, passwordHash string) (User, error) {
	var id int
	query := `
		INSERT INTO users (username, password_hash)
		VALUES ($1, $2)
		RETURNING id;
	`
	err := store.DB.QueryRowContext(ctx, query, username, passwordHash).Scan(&id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == pgerrcode.UniqueViolation {
				switch pgErr.ConstraintName {
				case "users_username_key":
					return User{}, ErrUserNameAlreadyExists
				default:
					return User{}, err
				}
			}
		}
		return User{}, err
	}
	return User{ID: id, UserName: username}, nil
}

// GetUserByUserName return User
func (store *PostgresStorage) GetUserByUserName(ctx context.Context, username string) (User, error) {
	query := `
		SELECT id, username, password_hash
		FROM users
		WHERE username = $1;
	`
	row := store.DB.QueryRowContext(ctx, query, username)
	var user User
	if err := row.Scan(
		&user.ID,
		&user.UserName,
		&user.PasswordHash,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return User{}, ErrUserNotFound
		}
		return User{}, err
	}

	return user, nil
}

// CreateOrder creates a new order and returns it
func (store *PostgresStorage) CreateOrder(ctx context.Context, orderNumber int, userID int) (Order, error) {
	var currentUserID int
	var isInsertionTime bool
	query := `
		INSERT INTO orders (order_number, user_id, uploaded_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (order_number) DO UPDATE SET order_number = orders.order_number
		RETURNING user_id, uploaded_at = $4
	`
	err := store.DB.QueryRowContext(ctx, query, orderNumber, userID).Scan(&currentUserID, isInsertionTime)
	if err != nil {
		return Order{}, err
	}
	if currentUserID != userID {
		return Order{}, ErrOrderCreatedByAnotherUser
	}
	if !isInsertionTime {
		return Order{}, ErrOrderAlreadyCreatedByUser
	}
	return Order{OrderNumber: orderNumber, UserID: userID}, nil
}

// GetOrdersByUser todo
func (store *PostgresStorage) GetOrdersByUser(ctx context.Context, userID int) ([]Order, error) {
	rows, err := store.DB.QueryContext(ctx, "SELECT id, order_number, user_id, status, accrual, uploaded_at FROM orders WHERE user_id = $1", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []Order
	for rows.Next() {
		var order Order
		if err := rows.Scan(&order.ID, &order.OrderNumber, &order.UserID, &order.Status, &order.Accrual, &order.UploadedAt); err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return orders, nil
}
