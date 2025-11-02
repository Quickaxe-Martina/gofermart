package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Quickaxe-Martina/gofermart/internal/config"
	"github.com/Quickaxe-Martina/gofermart/internal/handler"
	"github.com/Quickaxe-Martina/gofermart/internal/logger"
	"github.com/Quickaxe-Martina/gofermart/internal/repository"
	"github.com/Quickaxe-Martina/gofermart/internal/storage"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type setupRouterConfig struct {
	cfg         *config.Config
	store       storage.Storage
	orderWorker *repository.OrderWorkers
	logging     *zap.Logger
}

func setupRouter(setupCfg setupRouterConfig) *chi.Mux {
	r := chi.NewRouter()
	h := handler.NewHandler(setupCfg.cfg, setupCfg.store, setupCfg.orderWorker, setupCfg.logging)

	r.Use(h.RequestLogger)
	// r.Use(handler.GzipMiddleware)
	r.Route("/api/user", func(r chi.Router) {
		r.Post("/register", h.RegisterUser)
		r.Post("/login", h.LoginUser)
	})
	r.Route("/api/user/orders", func(r chi.Router) {
		r.With(h.UserMiddleware).Post("/", h.CreateOrder)
		r.With(h.UserMiddleware).Get("/", h.GetOrders)
		// r.Get("/urls", h.DeleteUserURLs)
	})
	r.Route("/api/user/balance", func(r chi.Router) {
		r.With(h.UserMiddleware).Get("/", h.GetUserBalance)
		r.With(h.UserMiddleware).Post("/withdraw", h.WithdrawUser)
	})
	r.Route("/api/user/withdrawals", func(r chi.Router) {
		r.With(h.UserMiddleware).Get("/", h.GetWithdrawalsByUser)
		// r.Get("/urls", h.DeleteUserURLs)
	})
	r.Route("/ping", func(r chi.Router) {
		r.Get("/", h.Ping)
	})
	return r
}

func main() {
	logging, err := logger.Initialize("info")
	if err != nil {
		log.Panic(err)
		return
	}

	cfg := config.NewConfig()
	store, err := storage.NewStorage(cfg, logging)
	if err != nil {
		logging.Error("Storage error", zap.Error(err))
		return
	}
	orderWorker, err := repository.NewOrderWorkers(repository.OrderWorkersConfig{
		Store:       store,
		NumWorkers:  cfg.NumWorkers,
		URL:         cfg.AccrualSystemAddress,
		PoolSize:    cfg.PullSize,
		PoolTimeout: time.Second * time.Duration(cfg.PoolTimeout),
		Logging:     logging,
	})

	if err != nil {
		logging.Error("Workers error", zap.Error(err))
		return
	}

	r := setupRouter(setupRouterConfig{
		cfg:         cfg,
		store:       store,
		orderWorker: orderWorker,
		logging:     logging,
	})

	logging.Info(cfg.DatabaseDsn)
	logging.Info("application is running")

	// Обработчик завершения (Ctrl+C, SIGTERM и т.п.)
	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
		<-ch
		orderWorker.Stop()
		store.Close()
		os.Exit(0)
	}()

	logging.Fatal("", zap.Error(http.ListenAndServe(cfg.RunAddr, r)))
}
