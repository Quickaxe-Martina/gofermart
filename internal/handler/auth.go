package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/Quickaxe-Martina/gofermart/internal/auth"
	"github.com/Quickaxe-Martina/gofermart/internal/logger"
	"github.com/Quickaxe-Martina/gofermart/internal/storage"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

type registerUserRequest struct {
	Login    string `json:"login" validate:"required,min=4,max=50,containsany=ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"`
	Password string `json:"password" validate:"required,min=8,max=50,containsany=ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789!@#$%^&*"`
}

// RegisterUser TODO
func (h *Handler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	var req registerUserRequest
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&req); err != nil {
		logger.Log.Info("cannot decode request JSON body", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	validate := validator.New(validator.WithRequiredStructEnabled())
	if err := validate.Struct(req); err != nil {
		logger.Log.Info("validation error", zap.Error(err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	passwordHash, err := auth.GenerateHash(req.Password)
	if err != nil {
		logger.Log.Error("", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	user, err := h.store.CreateUser(r.Context(), req.Login, passwordHash)
	if err != nil {
		if errors.Is(err, storage.ErrUserNameAlreadyExists) {
			http.Error(w, "UserName already taken", http.StatusConflict)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	if err := auth.GenerateAndSetTokenInCookie(w, h.cfg.SecretKey, time.Hour*time.Duration(h.cfg.TokenExp), user.ID, user.PasswordHash); err != nil {
		logger.Log.Error("failed to generate token", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)

}

// LoginUser TODO
func (h *Handler) LoginUser(w http.ResponseWriter, r *http.Request) {
	var req registerUserRequest
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&req); err != nil {
		logger.Log.Info("cannot decode request JSON body", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	validate := validator.New(validator.WithRequiredStructEnabled())
	if err := validate.Struct(req); err != nil {
		logger.Log.Info("validation error", zap.Error(err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, err := h.store.GetUserByUserName(r.Context(), req.Login)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			http.Error(w, "Invalid login/password pair", http.StatusUnauthorized)
		} else {
			logger.Log.Error("", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	if err := auth.VerifyPassowrd(req.Password, user.PasswordHash); err != nil {
		http.Error(w, "Invalid login/password pair", http.StatusUnauthorized)
		return
	}

	if err := auth.GenerateAndSetTokenInCookie(w, h.cfg.SecretKey, time.Hour*time.Duration(h.cfg.TokenExp), user.ID, user.PasswordHash); err != nil {
		logger.Log.Error("failed to generate token", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)

}
