package config

import (
	"flag"
	"os"
	"path/filepath"
	"time"

	"github.com/caarlos0/env/v11"
)

// generate:reset
type Config struct {
	ServerURL       string `env:"SERVER_ADDRESS"`
	BaseURL         string `env:"BASE_URL"`
	LogLevel        string `env:"LOG_LEVEL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	DatabaseAdr     string `env:"DATABASE_DSN"`
	JWTSecret       string `env:"JWT_SECRET"`
	AuditFile       string `env:"AUDIT_FILE"`
	AuditURL        string `env:"AUDIT_URL"`
	EnableHTTPS     *bool  `env:"ENABLE_HTTPS"`
	// Timeouts
	DeleteURLDelay     time.Duration
	DeleteURLStopAfter time.Duration
	HandlerCtxTimeout  time.Duration
	TestsTimeout       time.Duration
	WorkerTimeout      time.Duration
}

func NewConfig() (*Config, error) {
	var cfg Config
	err := env.Parse(&cfg)
	if err != nil {
		return nil, err
	}

	var serverURL, baseURL, logLevel, fileStoragePath, databaseAdr, auditFile, auditURL string
	var enableHTTPS bool
	var deleteDelay, deleteStopAfter, handlerTimeout, testsTimeout, workerTimeout int
	flag.StringVar(&serverURL, "a", ":8080", ":<port>")
	flag.StringVar(&baseURL, "b", "http://localhost:8080/", "Base URL for the API. Example: http://localhost:8080/")
	flag.StringVar(&logLevel, "l", "info", "Log level. Example: info, debug, error")
	flag.StringVar(&fileStoragePath, "f", "", "File storage path")
	flag.StringVar(&databaseAdr, "d", "", "Database connection string")
	flag.StringVar(&auditFile, "audit-file", "", "File path for audit")
	flag.StringVar(&auditURL, "audit-url", "", "URL for audit")
	flag.BoolVar(&enableHTTPS, "s", false, "Enable HTTPS")
	// Timeouts
	flag.IntVar(&deleteDelay, "dd", 100, "Deletion URL delay in milliseconds")
	flag.IntVar(&deleteStopAfter, "sa", 500, "Stop deletion after N milliseconds")
	flag.IntVar(&handlerTimeout, "ht", 3, "Timeout for handlers in seconds")
	flag.IntVar(&testsTimeout, "tt", 3, "Timeout for tests in seconds")
	flag.IntVar(&workerTimeout, "wt", 3, "Timeout for workers in seconds")

	// flag.StringVar(&fileStoragePath, "f", "/tmp/links.json", "File storage path")
	// flag.StringVar(&databaseAdr, "d", "postgres://postgres:12345@localhost:5432/shorted_links?sslmode=disable", "Database connection string")
	flag.Parse()

	// Timeouts
	if cfg.DeleteURLDelay == 0 {
		cfg.DeleteURLDelay = time.Duration(deleteDelay) * time.Millisecond
	}
	if cfg.DeleteURLStopAfter == 0 {
		cfg.DeleteURLStopAfter = time.Duration(deleteStopAfter) * time.Millisecond
	}
	if cfg.HandlerCtxTimeout == 0 {
		cfg.HandlerCtxTimeout = time.Duration(handlerTimeout) * time.Second
	}
	if cfg.TestsTimeout == 0 {
		cfg.TestsTimeout = time.Duration(testsTimeout) * time.Second
	}
	if cfg.WorkerTimeout == 0 {
		cfg.WorkerTimeout = time.Duration(workerTimeout) * time.Second
	}

	if cfg.ServerURL == "" {
		cfg.ServerURL = serverURL
	}
	if cfg.BaseURL == "" {
		cfg.BaseURL = baseURL
	}
	if cfg.LogLevel == "" {
		cfg.LogLevel = logLevel
	}
	if cfg.FileStoragePath == "" && fileStoragePath != "" {
		dir, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		cfg.FileStoragePath = filepath.Join(dir, fileStoragePath)
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
	if cfg.EnableHTTPS == nil {
		cfg.EnableHTTPS = &enableHTTPS
	}
	if cfg.JWTSecret == "" {
		cfg.JWTSecret = "secret"
	}

	return &cfg, nil
}
