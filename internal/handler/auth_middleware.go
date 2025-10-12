package handler

import (
	"context"
	"net/http"

	"github.com/Quickaxe-Martina/gofermart/internal/auth"
	"github.com/Quickaxe-Martina/gofermart/internal/logger"
	"github.com/Quickaxe-Martina/gofermart/internal/storage"
	"go.uber.org/zap"
)

type ctxKey string

const userKey = ctxKey("user")

// UserMiddleware get  middware
func (h *Handler) UserMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, err := auth.GetUserByCookie(r, h.cfg.SecretKey)
		if err != nil {
			logger.Log.Error("error get user", zap.Error(err))
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), userKey, user)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUser give user from context
func GetUser(ctx context.Context) storage.User {
	return ctx.Value(userKey).(storage.User)
}
