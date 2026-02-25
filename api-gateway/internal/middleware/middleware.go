package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/tronget/human-touch/shared/storage"
)

func JWT(secret []byte) func(next http.Handler) http.Handler {
	if len(secret) == 0 {
		slog.Error("missing JWT_SECRET in environment variables")
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// validate token if present and set X-User-ID header
			auth := r.Header.Get("Authorization")
			if auth == "" {
				http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
				return
			}

			authToken, ok := strings.CutPrefix(auth, "Bearer ")
			if !ok {
				http.Error(w, "Invalid Authorization header", http.StatusUnauthorized)
				return
			}

			token, err := jwt.Parse(authToken, func(t *jwt.Token) (any, error) {
				return secret, nil
			})
			if err != nil || !token.Valid {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				http.Error(w, "Invalid claims", http.StatusUnauthorized)
				return
			}

			if exp, ok := claims["exp"].(int64); !ok || exp < time.Now().Unix() {
				http.Error(w, "Token has expired", http.StatusUnauthorized)
				return
			}

			if uid, ok := claims["user_id"].(int64); !ok {
				http.Error(w, "Invalid user id", http.StatusUnauthorized)
				return
			} else {
				r.Header.Set("X-User-ID", fmt.Sprintf("%d", uid))
			}
			next.ServeHTTP(w, r)
		})
	}
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
			err := db.Get(&exists, "SELECT EXISTS(SELECT 1 FROM app_user WHERE id=$1)", uid)
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
