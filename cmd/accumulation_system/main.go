package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/Quickaxe-Martina/gofermart/internal/config"
	"github.com/Quickaxe-Martina/gofermart/internal/handler"
	"github.com/Quickaxe-Martina/gofermart/internal/logger"
	"github.com/Quickaxe-Martina/gofermart/internal/storage"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func setupRouter(cfg *config.Config, store storage.Storage) *chi.Mux {
	r := chi.NewRouter()
	h := handler.NewHandler(cfg, store)

	r.Use(logger.RequestLogger)
	r.Use(handler.GzipMiddleware)
	r.Route("/api/user", func(r chi.Router) {
		r.Post("/register", h.JSONGenerateURL)
		r.Post("/login", h.BatchGenerateURL)
	})
	// r.Route("/api/user", func(r chi.Router) {
	// 	r.With(h.GetOrCreateUserMiddleware).Get("/urls", h.GetUserURLs)
	// 	r.With(h.GetOrCreateUserMiddleware).Delete("/urls", h.DeleteUserURLs)
	// })
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

	if err := logger.Initialize("info"); err != nil {
		log.Panic(err)
	}
	r := setupRouter(cfg, store)

	// Обработчик завершения (Ctrl+C, SIGTERM и т.п.)
	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
		<-ch
		store.Close()
		os.Exit(0)
	}()

	logger.Log.Fatal("", zap.Error(http.ListenAndServe(cfg.RunAddr, r)))
}
