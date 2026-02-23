package config

type LogConfig struct {
	Env string `yaml:"env" env-default:"dev"`
}
