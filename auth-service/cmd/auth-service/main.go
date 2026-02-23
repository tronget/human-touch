package main

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/tronget/auth-service/internal/config"
	"github.com/tronget/auth-service/internal/domain/user"
	"github.com/tronget/auth-service/pkg/logx"
	"github.com/tronget/auth-service/pkg/storage"
)

func main() {
	cfg := config.MustLoad()

	logx.SetupLogger(cfg)

	dbx, err := sqlx.Connect("postgres", cfg.DSN)
	if err != nil {
		slog.Error("Error connecting to postgres database: " + err.Error())
	}
	db := storage.NewDB(dbx, slog.Default())
	userRepo := user.NewRepository(db)
	userService := user.NewService(userRepo)

	r := chi.NewRouter()

	r.Post("/register", user.RegisterHandler(userService))
	r.Post("/login", user.LoginHandler(userService, cfg))

	r.Group(func(r chi.Router) {
		r.Use(WithUID)
		r.Get("/me", user.MeHandler())
	})

	slog.Info("auth service listening on " + cfg.ServiceAddress)
	http.ListenAndServe(cfg.ServiceAddress, r)
}
