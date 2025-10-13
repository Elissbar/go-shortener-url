package config

type Config struct {
	ServerUrl  string
	BaseUrl    string
}

func New(serverUrl, baseUrl string) *Config {
	return &Config{
		ServerUrl:  serverUrl,
		BaseUrl:    baseUrl,
	}
}
