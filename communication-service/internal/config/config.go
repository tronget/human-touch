package config

import (
	"log/slog"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	ServiceAddress string `yaml:"server.address" env:"COMM_SERVICE_ADDRESS"`
	DSN            string `yaml:"postgre_dsn" env:"DATABASE_URL" required:"true"`
	JwtSecret      string `yaml:"jwt_secret" env:"JWT_SECRET" required:"true"`
	LogConfig
}

type LogConfig struct {
	Env string `yaml:"env" env-default:"dev"`
}

func MustLoad() *Config {
	configPath := os.Getenv("CONFIG_PATH")

	if configPath == "" {
		slog.Error("CONFIG_PATH environment variable is not set")
		panic("CONFIG_PATH environment variable is not set")
	}

	if _, err := os.Stat(configPath); err != nil {
		slog.Error("error opening configuration file", "error", err.Error())
		panic("error opening configuration file")
	}

	var cfg Config
	err := cleanenv.ReadConfig(configPath, &cfg)
	if err != nil {
		slog.Error("error reading configuration file", "error", err.Error())
		panic("error reading configuration file")
	}

	return &cfg
}
