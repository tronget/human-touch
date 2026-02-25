package main

import (
	"log/slog"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/tronget/human-touch/api-gateway/internal/config"
	"github.com/tronget/human-touch/api-gateway/internal/server"
	"github.com/tronget/human-touch/shared/logx"
	"github.com/tronget/human-touch/shared/storage"
)

func main() {
	cfg := config.MustLoad()

	logx.SetupLogger(cfg.Env)

	db, err := initDB(cfg)
	if err != nil {
		slog.Error("error initializing database", "err", err)
		return
	}
	defer db.X.Close()

	s := server.NewServer(cfg, db)
	if err := s.Run(); err != nil {
		slog.Error("error running server", "err", err)
		return
	}
}

func initDB(cfg *config.Config) (*storage.DB, error) {
	dbx, err := sqlx.Connect("postgres", cfg.DSN)
	if err != nil {
		return nil, err
	}
	db := storage.NewDB(dbx, slog.Default())
	return db, nil
}
