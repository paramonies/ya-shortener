// Package config implement functions for environment and project configs.
package config

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/caarlos0/env/v6"

	"github.com/paramonies/internal/store"
)

// Config contains all config variables for application.
type Config struct {
	SrvAddr       string `env:"SERVER_ADDRESS" envDefault:"localhost:8080"`
	BaseURL       string `env:"BASE_URL" envDefault:"http://localhost:8080"`
	FileStorePath string `env:"FILE_STORAGE_PATH"`
	//DatabaseDSN   string `env:"DATABASE_DSN" envDefault:"postgresql://postgres:123456@localhost:5432/shortener-api?connect_timeout=10&sslmode=disable"`
	DatabaseDSN    string `env:"DATABASE_DSN"`
	EnableHTTPS    *bool  `env:"ENABLE_HTTPS" envDefault:"false"`
	ConfigFileName string `env:"CONFIG"`
}

// JSONConfig for json config
type JSONConfig struct {
	BaseURL       string `json:"base_url"`
	SrvAddr       string `json:"server_address"`
	FileStorePath string `json:"file_storage_path"`
	DatabaseDSN   string `json:"database_dsn"`
	EnableHTTPS   bool   `json:"enable_https"`
}

// Init define Config variables from env variables or command args.
func (cfg *Config) Init() error {
	err := env.Parse(cfg)
	if err != nil {
		return err
	}

	flag.StringVar(&cfg.SrvAddr, "a", cfg.SrvAddr, "server host and port")
	flag.StringVar(&cfg.BaseURL, "b", cfg.BaseURL, "URL for making http request")
	flag.StringVar(&cfg.FileStorePath, "f", cfg.FileStorePath, "path to DB-file on disk")
	flag.StringVar(&cfg.DatabaseDSN, "d", cfg.DatabaseDSN, "database dsn")
	cfg.EnableHTTPS = flag.Bool("s", *cfg.EnableHTTPS, "enable HTTPS")
	flag.StringVar(&cfg.ConfigFileName, "c", cfg.ConfigFileName, "config file name")
	flag.StringVar(&cfg.ConfigFileName, "config", cfg.ConfigFileName, "config file name")

	flag.Parse()

	if cfg.ConfigFileName == "" {
		return nil
	}

	pwd, _ := os.Getwd()
	path := pwd + "/config/" + cfg.ConfigFileName
	fmt.Println("path ", path)
	byteValue, err := os.ReadFile(path)

	if err != nil {
		return err
	}
	var config JSONConfig

	err = json.Unmarshal(byteValue, &config)
	if err != nil {
		return err
	}

	if cfg.SrvAddr == "" {
		cfg.SrvAddr = config.SrvAddr
	}

	if cfg.BaseURL == "" {
		cfg.BaseURL = config.BaseURL
	}
	if cfg.FileStorePath == "" {
		if _, err := os.Stat(config.FileStorePath); !errors.Is(err, os.ErrNotExist) {
			cfg.FileStorePath = config.FileStorePath
		}
	}
	if cfg.DatabaseDSN == "" {
		cfg.DatabaseDSN = config.DatabaseDSN
	}
	if cfg.EnableHTTPS != nil {
		cfg.EnableHTTPS = &config.EnableHTTPS
	}

	return nil
}

// NewRepository create new repository.
func NewRepository(cfg *Config) (store.Repository, error) {
	var db store.Repository
	var err error
	if cfg.DatabaseDSN != "" {
		db, err = store.NewPostgresDB(cfg.DatabaseDSN)
		if err != nil {
			return nil, err
		}
	} else if cfg.FileStorePath != "" {
		db, err = store.NewFileDB(cfg.FileStorePath)
		if err != nil {
			return nil, err
		}
	} else {
		db = store.NewMapDB()
	}

	return db, nil
}
