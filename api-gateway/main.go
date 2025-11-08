package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
)

func proxyURL(target string) *httputil.ReverseProxy {
	u, _ := url.Parse(target)
	return httputil.NewSingleHostReverseProxy(u)
}

func jwtMiddleware(next http.Handler) http.Handler {
	secret := []byte(os.Getenv("JWT_SECRET"))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// validate token if present and set X-User-ID header
		auth := r.Header.Get("Authorization")
		if auth == "" {
			http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
			return
		}
		parts := strings.Split(auth, " ")
		if len(parts) == 2 && parts[0] == "Bearer" {
			token, err := jwt.Parse(parts[1], func(t *jwt.Token) (any, error) {
				return secret, nil
			})
			if err != nil || !token.Valid {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}
			if claims, ok := token.Claims.(jwt.MapClaims); ok {
				if uid, ok := claims["user_id"].(float64); ok {
					r.Header.Set("X-User-ID", fmt.Sprintf("%d", int64(uid)))
				}
			}
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	authProxy := http.StripPrefix("/auth", proxyURL("http://auth:8001"))
	commProxy := http.StripPrefix("/comm", proxyURL("http://comm:8002"))

	r := chi.NewRouter()

	r.Route("/auth", func(r chi.Router) { r.Handle("/*", authProxy) })

	r.Route("/comm", func(r chi.Router) {
		r.Use(jwtMiddleware)
		r.Handle("/*", commProxy)
	})

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	port := os.Getenv("API_GATEWAY_PORT")
	if port == "" {
		port = "8000"
	}
	log.Println("gateway on", port)
	http.ListenAndServe(":"+port, r)
}
