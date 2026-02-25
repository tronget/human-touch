package server

import (
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/go-chi/chi/v5"
	"github.com/tronget/human-touch/api-gateway/internal/config"
	"github.com/tronget/human-touch/api-gateway/internal/middleware"
	"github.com/tronget/human-touch/shared/storage"
)

type Server interface {
	Run() error
}

type server struct {
	cfg *config.Config
	db  *storage.DB
}

func NewServer(cfg *config.Config, db *storage.DB) Server {
	return &server{
		cfg: cfg,
		db:  db,
	}
}

func proxyURL(target string) *httputil.ReverseProxy {
	u, _ := url.Parse(target)
	return httputil.NewSingleHostReverseProxy(u)
}

func (s *server) Run() error {
	authProxy := http.StripPrefix("/auth", proxyURL(s.cfg.AuthServiceURL))
	commProxy := http.StripPrefix("/comm", proxyURL(s.cfg.CommServiceURL))

	r := chi.NewRouter()
	r.Use(
		middleware.CORS,
		middleware.RequestSizeLimit(10*1024*1024),
	)

	r.Route("/auth", func(r chi.Router) {
		r.Handle("/*", authProxy)
	})

	r.Route("/comm", func(r chi.Router) {
		r.Use(
			middleware.JWT([]byte(s.cfg.JwtSecret)),
			middleware.IsExistingUser(s.db),
		)
		r.Handle("/*", commProxy)
	})

	slog.Info("api gateway listening on " + s.cfg.ServiceAddress)
	return http.ListenAndServe(s.cfg.ServiceAddress, r)
}
