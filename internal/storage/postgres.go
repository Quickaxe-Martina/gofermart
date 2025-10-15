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
	return User{ID: int(id), UserName: username}, nil
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
