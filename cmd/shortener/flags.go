package main

import (
	"flag"
	"os"

	"github.com/Elissbar/go-shortener-url/internal/config"
)

func parseFlags() *config.Config {
	var serverUrl, baseUrl, logLevel string

	flag.StringVar(&serverUrl, "a", "localhost:8080", "<host>:<port>")
	flag.StringVar(&baseUrl, "b", "http://localhost:8080/", "Base URL for the API. Example: http://localhost:8080/")
	flag.StringVar(&logLevel, "l", "info", "Log level. Example: info, debug, error")
	flag.Parse()

	if osEnvServerAddr := os.Getenv("SERVER_ADDRESS"); osEnvServerAddr != "" {
		serverUrl = osEnvServerAddr
	}
	if osEnvBaseUrl := os.Getenv("BASE_URL"); osEnvBaseUrl != "" {
		baseUrl = osEnvBaseUrl
	}
	if osEnvLogLevel := os.Getenv("LOG_LEVEL"); osEnvLogLevel != "" {
		logLevel = osEnvLogLevel
	}

	return config.New(serverUrl, baseUrl, logLevel)
}
