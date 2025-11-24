package main

import (
	"flag"

	"github.com/Elissbar/go-shortener-url/internal/config"
	"github.com/caarlos0/env/v11"
)

func parseFlags() (*config.Config, error) {
	var cfg config.Config
	err := env.Parse(&cfg)
	if err != nil {
		return nil, err
	}

	var serverURL, baseURL, logLevel, fileStoragePath, databaseAdr string
	flag.StringVar(&serverURL, "a", ":8080", ":<port>")
	flag.StringVar(&baseURL, "b", "http://localhost:8080/", "Base URL for the API. Example: http://localhost:8080/")
	flag.StringVar(&logLevel, "l", "info", "Log level. Example: info, debug, error")
	flag.StringVar(&fileStoragePath, "f", "", "File storage path")
	// flag.StringVar(&databaseAdr, "d", "", "Database connection string")
	// flag.StringVar(&fileStoragePath, "f", "/tmp/links.json", "File storage path")
	flag.StringVar(&databaseAdr, "d", "postgres://postgres:12345@localhost:5432/shorted_links?sslmode=disable", "Database connection string")
	flag.Parse()

	if cfg.ServerURL == "" {
		cfg.ServerURL = serverURL
	}
	if cfg.BaseURL == "" {
		cfg.BaseURL = baseURL
	}
	if cfg.LogLevel == "" {
		cfg.LogLevel = logLevel
	}
	if cfg.FileStoragePath == "" {
		cfg.FileStoragePath = fileStoragePath
	}
	if cfg.DatabaseAdr == "" {
		cfg.DatabaseAdr = databaseAdr
	}

	return &cfg, nil
}
