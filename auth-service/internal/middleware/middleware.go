package middleware

import (
	"net/http"
	"strconv"

	"github.com/tronget/human-touch/auth-service/internal/userctx"
)

// func CORSMiddleware(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		w.Header().Set("Access-Control-Allow-Origin", "http://localhost")
// 		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
// 		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-User-ID, X-Request-ID")
// 		if r.Method == http.MethodOptions {
// 			w.WriteHeader(http.StatusNoContent)
// 			return
// 		}
// 		next.ServeHTTP(w, r)
// 	})
// }

func ExtractHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		userID := r.Header.Get("X-User-ID")
		if userID == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		_, err := strconv.ParseInt(userID, 10, 64)
		if err != nil {
			http.Error(w, "Invalid user ID", http.StatusUnauthorized)
			return
		}

		ctx = userctx.WithUserID(ctx, userID)

		if reqID := r.Header.Get("X-Request-ID"); reqID != "" {
			ctx = userctx.WithRequestID(ctx, reqID)
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
