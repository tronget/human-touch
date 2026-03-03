package middleware

import (
	"net/http"

	"github.com/tronget/human-touch/communication-service/internal/userctx"
)

func ExtractHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		if uid := r.Header.Get("X-User-ID"); uid != "" {
			ctx = userctx.WithUserID(ctx, uid)
		}

		if rid := r.Header.Get("X-Request-ID"); rid != "" {
			ctx = userctx.WithRequestID(ctx, rid)
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
