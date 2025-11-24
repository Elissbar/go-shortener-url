package config

type Config struct {
	ServerURL       string `env:"SERVER_ADDRESS"`
	BaseURL         string `env:"BASE_URL"`
	LogLevel        string `env:"LOG_LEVEL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	DatabaseAdr     string `env:"DATABASE_DSN"`
	JWTSecret       string `env:"JWT_SECRET"`
}
