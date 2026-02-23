package logx

import (
	"log"
	"log/slog"
	"os"

	"github.com/tronget/auth-service/internal/config"
)

const (
	EnvLocal = "local"
	EnvDev   = "dev"
	EnvProd  = "prod"
)

func SetupLogger(cfg *config.Config) {
	var handler slog.Handler

	switch cfg.Env {
	case EnvLocal:
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			AddSource: true,
			Level:     slog.LevelDebug,
		})
	case EnvDev:
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			AddSource: true,
			Level:     slog.LevelDebug,
		})
	case EnvProd:
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			AddSource: true,
			Level:     slog.LevelInfo,
		})
	default:
		log.Fatal("Incorrect env naming")
	}

	logger := slog.New(handler)

	slog.SetDefault(logger)
}
