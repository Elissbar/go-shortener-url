package config

type Config struct {
	ServerURL, BaseURL, LogLevel string
}

func New(serverURL, baseURL, logLevel string) *Config {
	return &Config{
		ServerURL:  serverURL,
		BaseURL:    baseURL,
		LogLevel: logLevel,
	}
}
