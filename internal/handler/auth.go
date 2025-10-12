package handler

import (
	"net/http"
	"encoding/json"

	"github.com/go-playground/validator/v10"
	"github.com/Quickaxe-Martina/gofermart/internal/logger"
	"go.uber.org/zap"
)

type registerUserRequest struct {
	Login string `json:"login" validate:"required,min=4,max=50,containsany=ABCDEFGHIJKLMNOPQRSTUVWXYZ,containsany=abcdefghijklmnopqrstuvwxyz,containsany=0123456789"`
	Password string `json:"password" validate:"required,min=8,max=50,containsany=ABCDEFGHIJKLMNOPQRSTUVWXYZ,containsany=abcdefghijklmnopqrstuvwxyz,containsany=0123456789,containsany=!@#$%^&*()_+\\-=\\[\\]{};':\"\\\\|,.<>\\/?~"`
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

}