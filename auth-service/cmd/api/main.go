package main

import (
	"log/slog"
	"time"

	_ "github.com/lib/pq"
	"github.com/tronget/human-touch/auth-service/internal/config"
	"github.com/tronget/human-touch/auth-service/internal/server"
	"github.com/tronget/human-touch/shared/logx"
	"github.com/tronget/human-touch/shared/storage"
)

const (
	dbMaxOpenConns    = 20
	dbMaxIdleConns    = 10
	dbConnMaxIdleTime = time.Minute * 30
)

func main() {
	cfg := config.MustLoad()

	logx.SetupLogger(cfg.Env)

	db, err := storage.InitDB(
		cfg.DSN,
		dbMaxOpenConns,
		dbMaxIdleConns,
		dbConnMaxIdleTime,
		slog.Default(),
	)
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
