package main

import (
	"log/slog"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/tronget/auth-service/internal/config"
	"github.com/tronget/auth-service/internal/server"
	"github.com/tronget/auth-service/pkg/logx"
	"github.com/tronget/auth-service/pkg/storage"
)

func main() {
	cfg := config.MustLoad()

	logx.SetupLogger(cfg)

	db, err := initDB(cfg)
	if err != nil {
		slog.Error("Error initializing database: " + err.Error())
		return
	}
	defer db.X.Close()

	s := server.NewServer(cfg, db)
	if err := s.Run(); err != nil {
		slog.Error("Error running server: " + err.Error())
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
