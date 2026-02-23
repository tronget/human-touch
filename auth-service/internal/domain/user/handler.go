package user

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/tronget/auth-service/internal/config"
	"github.com/tronget/auth-service/internal/constants"
	"github.com/tronget/auth-service/internal/dto"
)

func RegisterHandler(service Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var in dto.RegisterRequest
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			http.Error(w, "bad body", http.StatusBadRequest)
			return
		}

		err, statusCode := service.RegisterUser(
			in.Email,
			in.Password,
			in.Name,
		)
		if err != nil {
			http.Error(w, err.Error(), statusCode)
			return
		}

		w.WriteHeader(http.StatusCreated)
	}
}

func LoginHandler(service Service, cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var in dto.LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			http.Error(w, "bad body", http.StatusBadRequest)
			slog.Error("Bad login request", "error", err.Error())
			return
		}

		token, err, statusCode := service.LoginUser(
			in.Email,
			in.Password,
			[]byte(cfg.JwtSecret),
		)
		if err != nil {
			http.Error(w, err.Error(), statusCode)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		responseMsg := map[string]string{
			"token": "Bearer " + string(token),
		}
		json.NewEncoder(w).Encode(responseMsg)
	}
}

func MeHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid := r.Context().Value(constants.CtxUserID).(int64)
		msg := fmt.Sprintf("Your User ID: %d", uid)
		w.Write([]byte(msg))
	}
}
