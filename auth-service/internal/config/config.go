package config

import (
	"log"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	ServiceAddress string `yaml:"server.address" env:"AUTH_SERVICE_ADDRESS"`
	DSN            string `yaml:"postgre_dsn" env:"DATABASE_URL"`
	JwtSecret      string `yaml:"jwt_secret" env:"JWT_SECRET"`
	LogConfig
}

func MustLoad() *Config {
	configPath := os.Getenv("CONFIG_PATH")

	if configPath == "" {
		log.Fatal("CONFIG_PATH environment variable is not set")
	}

	if _, err := os.Stat(configPath); err != nil {
		log.Fatalf("error opening configuration file: %v", err)
	}

	var cfg Config
	err := cleanenv.ReadConfig(configPath, &cfg)
	if err != nil {
		log.Fatalf("error reading config: %v", err)
	}

	return &cfg
}
