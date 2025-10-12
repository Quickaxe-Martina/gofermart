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


// CreateUser creates a new user and returns it
func (store *PostgresStorage) CreateUser(ctx context.Context, username string, passowrHash string) (User, error) {
	var id int
	err := store.DB.QueryRowContext(ctx, "INSERT INTO users DEFAULT VALUES RETURNING id").Scan(&id)
	if err != nil {
		return User{}, err
	}
	return User{ID: int(id)}, nil
}
