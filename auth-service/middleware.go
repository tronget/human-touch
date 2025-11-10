package main

import (
	"context"
	"net/http"
	"strconv"
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

func WithUID() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			uid := r.Header.Get("X-User-ID")
			if uid == "" {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			if _, err := strconv.ParseInt(uid, 10, 64); err != nil {
				http.Error(w, "Invalid user ID", http.StatusUnauthorized)
				return
			}
			ctx := context.WithValue(r.Context(), CtxUserID, uid)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
