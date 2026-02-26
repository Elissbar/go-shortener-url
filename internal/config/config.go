package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/caarlos0/env/v11"
)

type ConfigFile struct {
	ServerAddr      string `json:"server_address"`
	BaseURL         string `json:"base_url"`
	FileStoragePath string `json:"file_storage_path"`
	DBDSN           string `json:"database_dsn"`
	EnableHTTPS     bool   `json:"enable_https"`
}

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
	ConfigFile      string `env:"CONFIG"`
	// Timeouts
	DeleteURLDelay     time.Duration
	DeleteURLStopAfter time.Duration
	HandlerCtxTimeout  time.Duration
	TestsTimeout       time.Duration
	WorkerTimeout      time.Duration
}

func loadEnv(cfg *Config) (*Config, error) {
	err := env.Parse(&cfg)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func loadFlags(cfg *Config) (*Config, error) {
	if cfg.ServerURL == "" {
		flag.StringVar(&cfg.ServerURL, "a", ":8080", ":<port>")
	}
	if cfg.BaseURL == "" {
		flag.StringVar(&cfg.BaseURL, "b", "http://localhost:8080/", "Base URL for the API. Example: http://localhost:8080/")
	}
	if cfg.LogLevel == "" {
		flag.StringVar(&cfg.LogLevel, "l", "info", "Log level. Example: info, debug, error")
	}
	if cfg.AuditFile == "" {
		flag.StringVar(&cfg.AuditFile, "audit-file", "", "File path for audit")
	}
	if cfg.AuditURL == "" {
		flag.StringVar(&cfg.AuditURL, "audit-url", "", "URL for audit")
	}
	if cfg.FileStoragePath == "" {
		dir, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("error get wd: %s", err)
		}
		var src string
		flag.StringVar(&src, "f", "", "File storage path")
		cfg.FileStoragePath = filepath.Join(dir, src)
		// flag.StringVar(&cfg.FileStoragePath, "f", "/tmp/links.json", "File storage path")
	}
	if cfg.DatabaseAdr == "" {
		flag.StringVar(&cfg.DatabaseAdr, "d", "", "Database connection string")
		// flag.StringVar(&cfg.DatabaseAdr, "d", "postgres://postgres:12345@localhost:5432/shorted_links?sslmode=disable", "Database connection string")
	}
	if cfg.JWTSecret == "" {
		cfg.JWTSecret = "secret"
	}
	if cfg.EnableHTTPS == nil {
		flag.BoolVar(cfg.EnableHTTPS, "s", false, "Enable HTTPS")
	}
	if cfg.ConfigFile == "" {
		flag.StringVar(&cfg.ConfigFile, "c", "", "Config File")
	}

	// Timeouts
	flag.DurationVar(&cfg.DeleteURLDelay, "dd", 100*time.Millisecond, "Deletion URL delay in milliseconds")
	flag.DurationVar(&cfg.DeleteURLStopAfter, "sa", 500*time.Millisecond, "Stop deletion after N milliseconds")
	flag.DurationVar(&cfg.HandlerCtxTimeout, "ht", 3*time.Second, "Timeout for handlers in seconds")
	flag.DurationVar(&cfg.TestsTimeout, "tt", 3*time.Second, "Timeout for tests in seconds")
	flag.DurationVar(&cfg.WorkerTimeout, "wt", 3*time.Second, "Timeout for workers in seconds")
	flag.Parse()

	return cfg, nil
}

func loadFile(cfg *Config) *Config {
	if cfg.ConfigFile == "" {
		return nil
	}

	data, err := os.ReadFile(cfg.ConfigFile)
	if err != nil {
		return nil
	}

	var cfgFile ConfigFile
	if err := json.Unmarshal(data, &cfgFile); err != nil {
		return nil
	}

	if cfg.ServerURL == "" {
		cfg.ServerURL = cfgFile.ServerAddr
	}
	if cfg.BaseURL == "" {
		cfg.BaseURL = cfgFile.BaseURL
	}
	if cfg.FileStoragePath == "" {
		cfg.FileStoragePath = cfgFile.FileStoragePath
	}
	if cfg.DatabaseAdr == "" {
		cfg.DatabaseAdr = cfgFile.DBDSN
	}
	if cfg.EnableHTTPS == nil {
		cfg.EnableHTTPS = &cfgFile.EnableHTTPS
	}

	return cfg
}

func NewConfig() (*Config, error) {
	var cfg *Config

	cfg, err := loadEnv(cfg)
	if err != nil {
		return nil, err
	}

	cfg, err = loadFlags(cfg)
	if err != nil {
		return nil, err
	}

	cfg = loadFile(cfg)

	return cfg, nil
}
