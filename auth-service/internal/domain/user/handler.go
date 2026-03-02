package user

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/tronget/human-touch/auth-service/internal/config"
	"github.com/tronget/human-touch/auth-service/internal/dto"
	"github.com/tronget/human-touch/auth-service/internal/userctx"
)

func RegisterHandler(service Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var in dto.RegisterRequest
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			http.Error(w, "bad body", http.StatusBadRequest)
			return
		}

		err, statusCode := service.RegisterUser(
			r.Context(),
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
			r.Context(),
			in.Email,
			in.Password,
			[]byte(cfg.JwtSecret),
		)
		if err != nil {
			http.Error(w, err.Error(), statusCode)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		responseMsg := map[string]string{
			"token": "Bearer " + string(token),
		}
		err = json.NewEncoder(w).Encode(responseMsg)
		if err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
	}
}

func MeHandler(service Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, ok := userctx.UserID(r.Context())
		if !ok || uid == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		userDto, err := service.GetUserByID(r.Context(), uid)
		if err != nil {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		err = json.NewEncoder(w).Encode(*userDto)
		if err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
	}
}
