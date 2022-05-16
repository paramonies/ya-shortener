package config

import (
	"flag"

	"github.com/caarlos0/env/v6"
)

type Config struct {
	SrvAddr       string `env:"SERVER_ADDRESS" envDefault:"localhost:8080"`
	BaseURL       string `env:"BASE_URL" envDefault:"http://localhost:8080"`
	FileStorePath string `env:"FILE_STORAGE_PATH"`
	//DatabaseDSN   string `env:"DATABASE_DSN" envDefault:"postgresql://postgres:123456@localhost:5432/shortener-api?connect_timeout=10&sslmode=disable"`
	DatabaseDSN string `env:"DATABASE_DSN"`
}

func (cfg *Config) Init() error {
	err := env.Parse(cfg)
	if err != nil {
		return err
	}

	flag.StringVar(&cfg.SrvAddr, "a", cfg.SrvAddr, "server host and port")
	flag.StringVar(&cfg.BaseURL, "b", cfg.BaseURL, "URL for making http request")
	flag.StringVar(&cfg.FileStorePath, "f", cfg.FileStorePath, "path to DB-file on disk")
	flag.StringVar(&cfg.DatabaseDSN, "d", cfg.DatabaseDSN, "database dsn")

	flag.Parse()

	return nil
}
