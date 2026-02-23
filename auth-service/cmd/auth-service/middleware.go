package main

import (
	"context"
	"net/http"
	"strconv"

	"github.com/tronget/auth-service/internal/constants"
)

func WithUID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uid := r.Header.Get("X-User-ID")
		if uid == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		v, err := strconv.ParseInt(uid, 10, 64)
		if err != nil {
			http.Error(w, "Invalid user ID", http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), constants.CtxUserID, v)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
