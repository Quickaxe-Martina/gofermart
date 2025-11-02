package storage

import (
	"context"
	"database/sql"
	"errors"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib" // driver

	"github.com/Quickaxe-Martina/gofermart/internal/config"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"go.uber.org/zap"
)

// PostgresStorage is DB implementation of the Storage interface
type PostgresStorage struct {
	DB      *sql.DB
	logging *zap.Logger
}

// NewPostgresStorage creates new PostgresStorage
func NewPostgresStorage(cfg *config.Config, logging *zap.Logger) (*PostgresStorage, error) {
	db, err := sql.Open("pgx", cfg.DatabaseDsn)
	if err != nil {
		panic(err)
	}
	if err := runMigrations(db, cfg.MigrationsPath); err != nil {
		return nil, err
	}
	store := &PostgresStorage{
		DB:      db,
		logging: logging,
	}
	return store, nil
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

// GetBalanceByUser get balance from DB
func (store *PostgresStorage) GetBalanceByUser(ctx context.Context, userID int) (UserBalance, error) {
	query := `
		SELECT balance, withdrawn
		FROM users
		WHERE id = $1;
	`
	row := store.DB.QueryRowContext(ctx, query, userID)
	var balance UserBalance
	if err := row.Scan(
		&balance.Balance,
		&balance.Withdrawn,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return UserBalance{}, ErrUserNotFound
		}
		return UserBalance{}, err
	}

	return balance, nil
}

// CreateOrder creates a new order and returns it
func (store *PostgresStorage) CreateOrder(ctx context.Context, orderNumber string, userID int) (Order, error) {
	var currentUserID int
	var isInsertionTime bool
	uploadedAt := time.Now().In(time.UTC)
	query := `
		INSERT INTO orders (order_number, user_id, uploaded_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (order_number) DO UPDATE SET order_number = orders.order_number
		RETURNING user_id, uploaded_at = $4;
	`
	err := store.DB.QueryRowContext(ctx, query, orderNumber, userID, uploadedAt, uploadedAt).Scan(&currentUserID, &isInsertionTime)
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

// GetOrdersByUser get user orders from DB
func (store *PostgresStorage) GetOrdersByUser(ctx context.Context, userID int) ([]Order, error) {
	query := `
		SELECT id, order_number, user_id, status, accrual, uploaded_at 
		FROM orders 
		WHERE user_id = $1 
		ORDER BY uploaded_at DESC;
	`
	rows, err := store.DB.QueryContext(ctx, query, userID)
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

// WithdrawUser withraw user
func (store *PostgresStorage) WithdrawUser(ctx context.Context, userID int, sum float64, orderNumber string) error {
	tx, err := store.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var balance, withdrawn float64
	row := tx.QueryRowContext(ctx, `SELECT balance, withdrawn FROM users WHERE id = $1 FOR UPDATE`, userID)
	if err := row.Scan(&balance, &withdrawn); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrUserNotFound
		}
		store.logging.Error("cannot lock user row", zap.Error(err))
		return err
	}

	if balance < sum {
		return ErrLowBalance
	}

	newBalance := balance - sum
	newWithdrawn := withdrawn + sum

	_, err = tx.ExecContext(ctx, `
        UPDATE users
        SET balance = $1, withdrawn = $2
        WHERE id = $3
    `, newBalance, newWithdrawn, userID)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `
        INSERT INTO withdrawals (user_id, order_number, sum, created_at)
        VALUES ($1, $2, $3, NOW())
    `, userID, orderNumber, sum)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// GetWithdrawalsByUser get withdrawals from DB
func (store *PostgresStorage) GetWithdrawalsByUser(ctx context.Context, userID int) ([]Withdrawal, error) {
	query := `
		SELECT order_number, sum, created_at
		FROM withdrawals
		WHERE user_id = $1
		ORDER BY created_at DESC;
	`
	rows, err := store.DB.QueryContext(ctx, query, userID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var withdrawals []Withdrawal
	for rows.Next() {
		var withdrawal Withdrawal
		if err := rows.Scan(&withdrawal.OrderNumber, &withdrawal.Sum, &withdrawal.CreatedAt); err != nil {
			return nil, err
		}
		withdrawals = append(withdrawals, withdrawal)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return withdrawals, nil
}

// AccrueUser ещвщ
func (store *PostgresStorage) AccrueUser(ctx context.Context, userID int, sum float64, status string, orderNumber string) error {
	tx, err := store.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var balance float64
	row := tx.QueryRowContext(ctx, `SELECT balance FROM users WHERE id = $1 FOR UPDATE`, userID)
	if err := row.Scan(&balance); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrUserNotFound
		}
		store.logging.Error("cannot lock user row", zap.Error(err))
		return err
	}

	newBalance := balance + sum

	_, err = tx.ExecContext(ctx, `
        UPDATE users
        SET balance = $1
        WHERE id = $2
    `, newBalance, userID)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `
        UPDATE orders
		SET status = $1, accrual = $2
		WHERE order_number = $3;
    `, status, sum, orderNumber)
	if err != nil {
		return err
	}

	return tx.Commit()
}
