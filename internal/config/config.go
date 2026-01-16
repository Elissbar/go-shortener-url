package config

import (
	"flag"
	"os"
	"path/filepath"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	ServerURL       string `env:"SERVER_ADDRESS"`
	BaseURL         string `env:"BASE_URL"`
	LogLevel        string `env:"LOG_LEVEL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	DatabaseAdr     string `env:"DATABASE_DSN"`
	JWTSecret       string `env:"JWT_SECRET"`
	AuditFile       string `env:"AUDIT_FILE"`
	AuditURL        string `env:"AUDIT_URL"`
}

func NewConfig() (*Config, error) {
	var cfg Config
	err := env.Parse(&cfg)
	if err != nil {
		return nil, err
	}

	var serverURL, baseURL, logLevel, fileStoragePath, databaseAdr, auditFile, auditURL string
	flag.StringVar(&serverURL, "a", ":8080", ":<port>")
	flag.StringVar(&baseURL, "b", "http://localhost:8080/", "Base URL for the API. Example: http://localhost:8080/")
	flag.StringVar(&logLevel, "l", "info", "Log level. Example: info, debug, error")
	flag.StringVar(&fileStoragePath, "f", "", "File storage path")
	flag.StringVar(&databaseAdr, "d", "", "Database connection string")
	flag.StringVar(&auditFile, "audit-file", "", "File path for audit")
	flag.StringVar(&auditURL, "audit-url", "", "URL for audit")

	// flag.StringVar(&fileStoragePath, "f", "/tmp/links.json", "File storage path")
	// flag.StringVar(&databaseAdr, "d", "postgres://postgres:12345@localhost:5432/shorted_links?sslmode=disable", "Database connection string")
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
	if cfg.AuditFile == "" && auditFile != "" {
		dir, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		cfg.AuditFile = filepath.Join(dir, auditFile)
	}
	if cfg.AuditURL == "" {
		cfg.AuditURL = auditURL
	}

	return &cfg, nil
}
