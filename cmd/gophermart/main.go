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

func setupRouter(cfg *config.Config, store storage.Storage, orderWorker *repository.OrderWorkers) *chi.Mux {
	r := chi.NewRouter()
	h := handler.NewHandler(cfg, store, orderWorker)

	r.Use(logger.RequestLogger)
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
	cfg := config.NewConfig()
	store, err := storage.NewStorage(cfg)
	if err != nil {
		logger.Log.Error("Storage error", zap.Error(err))
	}
	orderWorker := repository.NewOrderWorkers(store, cfg.NumWorkers, cfg.AccuralSystemAddress, cfg.PullSize, time.Second*time.Duration(cfg.PoolTimeout))

	if err := logger.Initialize("info"); err != nil {
		log.Panic(err)
	}
	r := setupRouter(cfg, store, orderWorker)
	logger.Log.Info(cfg.DatabaseDsn)
	logger.Log.Info("application is running")

	// Обработчик завершения (Ctrl+C, SIGTERM и т.п.)
	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
		<-ch
		orderWorker.Stop()
		store.Close()
		os.Exit(0)
	}()

	logger.Log.Fatal("", zap.Error(http.ListenAndServe(cfg.RunAddr, r)))
}
