package server

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/tronget/human-touch/auth-service/internal/config"
	"github.com/tronget/human-touch/auth-service/internal/domain/user"
	"github.com/tronget/human-touch/auth-service/internal/middleware"
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

func (s *server) Run() error {
	userRepo := user.NewRepository(s.db)
	userService := user.NewService(userRepo)

	r := chi.NewRouter()

	r.Post("/register", user.RegisterHandler(userService))
	r.Post("/login", user.LoginHandler(userService, s.cfg))

	r.Group(func(r chi.Router) {
		r.Use(middleware.WithUID)
		r.Get("/me", user.MeHandler())
	})

	slog.Info("auth service listening on " + s.cfg.ServiceAddress)
	return http.ListenAndServe(s.cfg.ServiceAddress, r)
}
