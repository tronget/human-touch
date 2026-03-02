package middleware

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"github.com/tronget/human-touch/shared/jwtx"
	"github.com/tronget/human-touch/shared/storage"
)

func JWT(secret []byte) func(next http.Handler) http.Handler {
	if len(secret) == 0 {
		slog.Error("missing JWT_SECRET in environment variables")
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := r.Header.Get("Authorization")
			if token == "" {
				token = r.URL.Query().Get("token")
			}
			if token == "" {
				http.Error(w, "Missing Authorization token", http.StatusUnauthorized)
				return
			}

			uid, err := jwtx.ValidateAndExtractUserID(token, secret)
			if err != nil {
				http.Error(w, "Invalid token: "+err.Error(), http.StatusUnauthorized)
				return
			}
			r.Header.Set("X-User-ID", fmt.Sprintf("%d", uid))
			next.ServeHTTP(w, r)
		})
	}
}

func WithRequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := uuid.New().String()
		r.Header.Set("X-Request-ID", requestID)
		w.Header().Set("X-Request-ID", requestID)

		next.ServeHTTP(w, r)
	})
}

func RequestSizeLimit(limit int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, limit)
			next.ServeHTTP(w, r)
		})
	}
}

func IsExistingUser(db *storage.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			uid := r.Header.Get("X-User-ID")
			if uid == "" {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			var exists bool
			err := db.Get(&exists, "SELECT EXISTS(SELECT 1 FROM users WHERE id=$1)", uid)
			if err != nil {
				http.Error(w, "Internal error", http.StatusInternalServerError)
				slog.Error("user existence check failed", "err", err)
				return
			}

			if !exists {
				http.Error(w, "User does not exist", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
