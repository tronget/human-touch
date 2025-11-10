package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/tronget/api-gateway/middleware"
)

func proxyURL(target string) *httputil.ReverseProxy {
	u, _ := url.Parse(target)
	return httputil.NewSingleHostReverseProxy(u)
}

func main() {
	authProxy := http.StripPrefix("/auth", proxyURL("http://auth:8001"))
	commProxy := http.StripPrefix("/comm", proxyURL("http://comm:8002"))

	r := chi.NewRouter()

	r.Route("/auth", func(r chi.Router) {
		r.Handle("/login", authProxy)
		r.Handle("/register", authProxy)

		// remaining auth routes protected by JWT middleware
		r.Group(func(r chi.Router) {
			r.Use(middleware.JWT)
			r.Handle("/*", authProxy)
		})
	})

	r.Route("/comm", func(r chi.Router) {
		r.Use(middleware.JWT)
		r.Handle("/*", commProxy)
	})

	port := os.Getenv("API_GATEWAY_PORT")
	if port == "" {
		port = "8000"
	}
	log.Println("gateway on", port)
	http.ListenAndServe(":"+port, r)
}
