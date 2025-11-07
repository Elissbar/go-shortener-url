package main

import (
	"flag"
	"os"

	"github.com/Elissbar/go-shortener-url/internal/config"
)

func parseFlags() *config.Config {
	cfg := &config.Config{}

	flag.StringVar(&cfg.ServerURL, "a", "localhost:8080", "<host>:<port>")
	flag.StringVar(&cfg.BaseURL, "b", "http://localhost:8080/", "Base URL for the API. Example: http://localhost:8080/")
	flag.StringVar(&cfg.LogLevel, "l", "info", "Log level. Example: info, debug, error")
	flag.StringVar(&cfg.FileStoragePath, "f", "/tmp/links.json", "File storage path")
	flag.Parse()

	if osEnvServerAddr := os.Getenv("SERVER_ADDRESS"); osEnvServerAddr != "" {
		cfg.ServerURL = osEnvServerAddr
	}
	if osEnvBaseUrl := os.Getenv("BASE_URL"); osEnvBaseUrl != "" {
		cfg.BaseURL = osEnvBaseUrl
	}
	if osEnvLogLevel := os.Getenv("LOG_LEVEL"); osEnvLogLevel != "" {
		cfg.LogLevel = osEnvLogLevel
	}
	if envPath := os.Getenv("FILE_STORAGE_PATH"); envPath != "" {
		cfg.FileStoragePath = envPath
	}

	return cfg
}
