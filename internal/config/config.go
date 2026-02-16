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

	var (
		serverURL, baseURL, logLevel, fileStorage, dbURI, auditFile, auditURL string
		deletionDelay, stopAfter, handlerTmt, testsTmt, workerTmt         int
		https                                                             bool
	)

	flag.StringVar(&serverURL, "a", ":8080", ":<port>")
	flag.StringVar(&baseURL, "b", "http://localhost:8080/", "Base URL for the API. Example: http://localhost:8080/")
	flag.StringVar(&logLevel, "l", "info", "Log level. Example: info, debug, error")
	flag.StringVar(&auditFile, "audit-file", "", "File path for audit")
	flag.StringVar(&auditURL, "audit-url", "", "URL for audit")
	flag.StringVar(&fileStorage, "f", "", "File storage path")
	flag.StringVar(&dbURI, "d", "", "Database connection string")
	// flag.StringVar(&fileStorage, "f", "/tmp/links.json", "File storage path")
	// flag.StringVar(&dbURI, "d", "postgres://postgres:12345@localhost:5432/shorted_links?sslmode=disable", "Database connection string")
	flag.BoolVar(&https, "s", false, "Enable HTTPS")
	// Timeouts
	flag.IntVar(&deletionDelay, "dd", 100, "Deletion URL delay in milliseconds")
	flag.IntVar(&stopAfter, "sa", 500, "Stop deletion after N milliseconds")
	flag.IntVar(&handlerTmt, "ht", 3, "Timeout for handlers in seconds")
	flag.IntVar(&testsTmt, "tt", 3, "Timeout for tests in seconds")
	flag.IntVar(&workerTmt, "wt", 3, "Timeout for workers in seconds")
	flag.Parse()

	applyIfEmpty(&cfg.ServerURL, serverURL)
	applyIfEmpty(&cfg.BaseURL, baseURL)
	applyIfEmpty(&cfg.LogLevel, logLevel)
	applyIfEmpty(&cfg.DatabaseAdr, dbURI)
	applyIfEmpty(&cfg.AuditURL, auditURL)
	
	applyPathIfEmpty(&cfg.FileStoragePath, fileStorage)
	applyPathIfEmpty(&cfg.AuditFile, auditFile)

	applyDurationMs(&cfg.DeleteURLDelay, deletionDelay)
	applyDurationMs(&cfg.DeleteURLStopAfter, stopAfter)
	applyDurationSec(&cfg.HandlerCtxTimeout, handlerTmt)
	applyDurationSec(&cfg.TestsTimeout, testsTmt)
	applyDurationSec(&cfg.WorkerTimeout, workerTmt)

	if cfg.EnableHTTPS == nil {
		cfg.EnableHTTPS = &https
	}
	if cfg.JWTSecret == "" {
		cfg.JWTSecret = "secret"
	}

	return &cfg, nil
}

func applyIfEmpty(dst *string, src string) {
	if *dst == "" && src != "" {
		*dst = src
	}
}

func applyPathIfEmpty(dst *string, src string) {
	if *dst == "" && src != "" {
		dir, _ := os.Getwd()
		*dst = filepath.Join(dir, src)
	}
}

func applyDurationMs(dst *time.Duration, src int) {
	if *dst == 0 && src != 0 {
		*dst = time.Duration(src) * time.Millisecond
	}
}

func applyDurationSec(dst *time.Duration, src int) {
	if *dst == 0 && src != 0 {
		*dst = time.Duration(src) * time.Second
	}
}