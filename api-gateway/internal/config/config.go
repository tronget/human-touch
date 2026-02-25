package config

import (
	"log"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	ServiceAddress string `yaml:"server.address" env:"API_GATEWAY_ADDRESS" env-default:":8000"`
	DSN            string `yaml:"postgre_dsn" env:"DATABASE_URL" required:"true"`
	JwtSecret      string `yaml:"jwt_secret" env:"JWT_SECRET" required:"true"`
	AuthServiceURL string `yaml:"auth_service_url" env:"AUTH_SERVICE_URL" env-default:"http://auth:8001"`
	CommServiceURL string `yaml:"comm_service_url" env:"COMM_SERVICE_URL" env-default:"http://comm:8002"`
	LogConfig
}

type LogConfig struct {
	Env string `yaml:"env" env-default:"dev"`
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
