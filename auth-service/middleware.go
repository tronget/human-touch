package main

import (
	"context"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type ctxKey string

const (
	CtxDBKey  ctxKey = "db"
	CtxUserID ctxKey = "user_id"
)

func WithDB(db *DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), CtxDBKey, db)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func JWTAuth(next http.Handler) http.Handler {
	secret := []byte(os.Getenv("JWT_SECRET"))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth == "" {
			http.Error(w, "missing auth", http.StatusUnauthorized)
			return
		}
		parts := strings.Split(auth, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "invalid auth header", http.StatusUnauthorized)
			return
		}
		token, err := jwt.Parse(parts[1], func(t *jwt.Token) (any, error) {
			return secret, nil
		})
		if err != nil || !token.Valid {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, "invalid claims", http.StatusUnauthorized)
			return
		}
		uidFloat, ok := claims["user_id"].(float64)
		if !ok {
			http.Error(w, "invalid user id", http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), CtxUserID, int64(uidFloat))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
