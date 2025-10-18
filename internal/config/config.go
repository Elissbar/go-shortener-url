package config

type Config struct {
	ServerURL  string
	BaseURL   string
}

func New(serverURL, baseURL string) *Config {
	return &Config{
		ServerURL:  serverURL,
		BaseURL:    baseURL,
	}
}
